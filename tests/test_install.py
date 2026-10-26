"""Installer regression checks using disposable directories and a fake Docker CLI."""

import os
from pathlib import Path
import subprocess
import tempfile
import unittest


INSTALLER = Path(__file__).resolve().parents[1] / "install.sh"


class InstallerTests(unittest.TestCase):
    def setUp(self):
        self.tmp = tempfile.TemporaryDirectory(prefix="nodalibre-install-test-")
        self.addCleanup(self.tmp.cleanup)
        self.root = Path(self.tmp.name)
        self.profile = self.root / "profile"
        self.profile.mkdir()
        self.bin = self.root / "bin"
        self.bin.mkdir()
        self.log = self.root / "docker.log"
        docker = self.bin / "docker"
        docker.write_text(
            '#!/bin/sh\n'
            'case "$*" in "compose version"|"info") exit 0 ;; esac\n'
            'printf "%s|%s\\n" "$PWD" "$*" >> "$INSTALL_TEST_LOG"\n'
            'if [ "$*" = "compose pull" ]; then exit "${INSTALL_TEST_PULL_EXIT:-0}"; fi\n'
        )
        docker.chmod(0o755)
        self.env = {
            key: value for key, value in os.environ.items()
            if not key.startswith(("NODALIBRE_", "LIBRE_CLAUDE_", "COMPOSE_", "GITHUB_OAUTH_"))
            and key not in ("PUBLIC_URL", "MISTRAL_API_KEYS")
        }
        self.env.update(
            HOME=str(self.profile),
            PATH=str(self.bin) + os.pathsep + self.env["PATH"],
            INSTALL_TEST_LOG=str(self.log),
        )

    def run_installer(self, *args, env=None):
        return subprocess.run(
            ["sh", str(INSTALLER), *args], cwd=self.root,
            env=dict(self.env, **(env or {})), text=True, capture_output=True,
        )

    def existing(self, directory, name="docker-compose.yml"):
        directory.mkdir(parents=True, exist_ok=True)
        compose = directory / name
        compose.write_text(
            'services:\n  libre-claude:\n    image: liberchat/libre-claude:latest\n'
            '    ports: ["9123:80"]\n'
            '    volumes: ["historical_data:/var/www/html/data"]\n'
            'volumes:\n  historical_data:\n    external: true\n'
        )
        config = directory / ".env"
        config.write_text("MISTRAL_API_KEYS=example-fixture\nPUBLIC_URL=https://example.invalid\n")
        return {path: path.read_bytes() for path in (compose, config)}

    def assert_preserved(self, files):
        for path, content in files.items():
            self.assertEqual(path.read_bytes(), content)

    def test_new_installation_uses_historical_image(self):
        result = self.run_installer("--yes", "--no-start")
        self.assertEqual(result.returncode, 0, result.stderr)
        target = self.profile / "nodalibre"
        self.assertIn("image: liberchat/libre-claude:latest", (target / "docker-compose.yml").read_text())
        self.assertFalse(self.log.exists())

    def test_legacy_environment_variables_still_work(self):
        target = self.root / "old install with spaces"
        result = self.run_installer(env={
            "LIBRE_CLAUDE_DIR": str(target), "LIBRE_CLAUDE_PORT": "9123",
            "LIBRE_CLAUDE_IMAGE": "example/image:custom", "LIBRE_CLAUDE_YES": "1",
            "LIBRE_CLAUDE_NO_START": "1", "LIBRE_CLAUDE_INSTALL_DOCKER": "0",
        })
        self.assertEqual(result.returncode, 0, result.stderr)
        compose = (target / "docker-compose.yml").read_text()
        self.assertIn('"9123:80"', compose)
        self.assertIn("image: example/image:custom", compose)
        self.assertFalse(self.log.exists())

    def test_new_environment_and_explicit_options_have_priority(self):
        target = self.root / "explicit"
        result = self.run_installer("--yes", "--no-start", "--dir", str(target), "--port", "9001", env={
            "LIBRE_CLAUDE_DIR": str(self.root / "old"), "NODALIBRE_DIR": str(self.root / "new"),
            "LIBRE_CLAUDE_IMAGE": "old/image:tag", "NODALIBRE_IMAGE": "liberchat/nodalibre:latest",
            "LIBRE_CLAUDE_PORT": "9002", "NODALIBRE_PORT": "9003",
        })
        self.assertEqual(result.returncode, 0, result.stderr)
        content = (target / "docker-compose.yml").read_text()
        self.assertIn('"9001:80"', content)
        self.assertIn("image: liberchat/nodalibre:latest", content)
        self.assertFalse((self.root / "old").exists())
        self.assertFalse((self.root / "new").exists())

    def test_existing_compose_variants_are_preserved_and_updated(self):
        for name in ("compose.yaml", "compose.yml", "docker-compose.yaml", "docker-compose.yml"):
            with self.subTest(name=name):
                target = self.root / name / "existing instance"
                files = self.existing(target, name)
                override = target / "compose.override.yaml"
                override.write_text("# existing overrides\n")
                files[override] = override.read_bytes()
                result = self.run_installer("--yes", "--dir", str(target), "--port", "9999")
                self.assertEqual(result.returncode, 0, result.stderr)
                self.assert_preserved(files)
                self.assertEqual(self.log.read_text().splitlines()[-2:], [
                    f"{target}|compose pull", f"{target}|compose up -d",
                ])
                self.assertFalse((target / "data").exists())

    def test_default_directory_reuses_legacy_installation(self):
        target = self.profile / "libre-claude"
        files = self.existing(target)
        result = self.run_installer("--yes", "--no-start")
        self.assertEqual(result.returncode, 0, result.stderr)
        self.assert_preserved(files)
        self.assertFalse((self.profile / "nodalibre").exists())

    def test_existing_new_directory_takes_priority(self):
        old = self.existing(self.profile / "libre-claude")
        new = self.existing(self.profile / "nodalibre")
        result = self.run_installer("--yes")
        self.assertEqual(result.returncode, 0, result.stderr)
        self.assert_preserved(old)
        self.assert_preserved(new)
        self.assertIn(f"{self.profile / 'nodalibre'}|compose pull", self.log.read_text())

    def test_partial_installation_keeps_existing_env(self):
        target = self.root / "partial"
        target.mkdir()
        config = target / ".env"
        config.write_text("MISTRAL_API_KEYS=existing-fixture\n")
        original = config.read_bytes()
        result = self.run_installer("--yes", "--no-start", "--dir", str(target))
        self.assertEqual(result.returncode, 0, result.stderr)
        self.assertEqual(config.read_bytes(), original)

    def test_dry_run_writes_nothing(self):
        result = self.run_installer("--yes", "--dry-run")
        self.assertEqual(result.returncode, 0, result.stderr)
        self.assertFalse((self.profile / "nodalibre").exists())
        self.assertFalse(self.log.exists())

    def test_failed_pull_never_starts_container(self):
        target = self.root / "existing"
        files = self.existing(target)
        result = self.run_installer("--yes", "--dir", str(target), env={"INSTALL_TEST_PULL_EXIT": "17"})
        self.assertEqual(result.returncode, 17)
        self.assert_preserved(files)
        self.assertEqual(self.log.read_text().splitlines(), [f"{target}|compose pull"])


if __name__ == "__main__":
    unittest.main()
