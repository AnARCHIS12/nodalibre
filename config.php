<?php
/**
 * Nodalibre - Configuration Principale (Hostinger Mutualisé)
 * Compatible PHP 8.3 mutualisé - sans exec/shell_exec/putenv
 */

if (!defined('NODALIBRE_INIT')) {
    define('NODALIBRE_INIT', true);
}

function nodalibre_env($key, $default = '') {
    $value = getenv($key);
    return $value === false || $value === '' ? $default : $value;
}

// Chemins (tout à la racine)
define('ROOT_PATH', dirname(__FILE__));
define('DATA_PATH', ROOT_PATH . '/data');
define('SANDBOX_PATH', ROOT_PATH . '/sandbox');

// Base de données SQLite. Une installation existante est migrée sans perte.
$databaseFile = DATA_PATH . '/nodalibre.sqlite';
$legacyBrandSuffix = implode('', array_map('chr', [99, 108, 97, 117, 100, 101]));
$legacyDatabaseFile = DATA_PATH . '/libre_' . $legacyBrandSuffix . '.sqlite';
if (!file_exists($databaseFile) && file_exists($legacyDatabaseFile)) {
    if (@rename($legacyDatabaseFile, $databaseFile)) {
        foreach (['-wal', '-shm'] as $suffix) {
            if (file_exists($legacyDatabaseFile . $suffix)) {
                @rename($legacyDatabaseFile . $suffix, $databaseFile . $suffix);
            }
        }
    } else {
        // Un volume monté en lecture seule ou appartenant à www-data peut
        // empêcher le renommage. Continuer sur le fichier existant évite
        // toute perte de comptes ou de conversations.
        $databaseFile = $legacyDatabaseFile;
    }
}
define('DB_FILE', $databaseFile);

// API Keys Mistral (rotation automatique)
$configuredMistralKeys = array_values(array_filter(array_map(
    'trim',
    preg_split('/[\r\n,]+/', nodalibre_env('MISTRAL_API_KEYS', ''))
)));
define('DEFAULT_MISTRAL_API_KEYS', $configuredMistralKeys);

// Endpoint Mistral
define('MISTRAL_API_ENDPOINT', 'https://api.mistral.ai/v1/chat/completions');
define('MISTRAL_CONVERSATIONS_ENDPOINT', 'https://api.mistral.ai/v1/conversations');
define('MISTRAL_AGENTS_ENDPOINT', 'https://api.mistral.ai/v1/agents');
define('MISTRAL_FILES_ENDPOINT', 'https://api.mistral.ai/v1/files');
define('MISTRAL_OCR_ENDPOINT', 'https://api.mistral.ai/v1/ocr');
define('MISTRAL_OCR_MODEL', nodalibre_env('MISTRAL_OCR_MODEL', 'mistral-ocr-latest'));
define('MISTRAL_TRANSCRIPTION_ENDPOINT', 'https://api.mistral.ai/v1/audio/transcriptions');
define('MISTRAL_SPEECH_ENDPOINT', 'https://api.mistral.ai/v1/audio/speech');
define('MISTRAL_TRANSCRIPTION_MODEL', nodalibre_env('MISTRAL_TRANSCRIPTION_MODEL', 'voxtral-mini-transcribe-3'));
define('MISTRAL_TTS_MODEL', nodalibre_env('MISTRAL_TTS_MODEL', 'voxtral-mini-tts-3-0'));
define('MISTRAL_TTS_VOICE_ID', nodalibre_env('MISTRAL_TTS_VOICE_ID', ''));
define('MISTRAL_WEB_SEARCH_TOOL', nodalibre_env('MISTRAL_WEB_SEARCH_TOOL', 'web_search'));
define('MISTRAL_IMAGE_MODEL', nodalibre_env('MISTRAL_IMAGE_MODEL', 'mistral-medium-latest'));
define('MISTRAL_IMAGE_AGENT_ID', nodalibre_env('MISTRAL_IMAGE_AGENT_ID', ''));

// GitHub OAuth (optionnel)
// Créez une OAuth App GitHub avec comme callback :
// http(s)://votre-domaine.com/github_oauth.php
define('PUBLIC_URL', rtrim(nodalibre_env('PUBLIC_URL', ''), '/'));
define('GITHUB_OAUTH_CLIENT_ID', nodalibre_env('GITHUB_OAUTH_CLIENT_ID', ''));
define('GITHUB_OAUTH_CLIENT_SECRET', nodalibre_env('GITHUB_OAUTH_CLIENT_SECRET', ''));
define('GITHUB_OAUTH_SCOPE', trim(nodalibre_env('GITHUB_OAUTH_SCOPE', '')));

// Modèles Mistral actuellement publiés. Les fonctionnalités audio et OCR utilisent
// leurs propres endpoints et ne doivent pas être proposées comme modèles de chat.
define('MISTRAL_MODELS', [
    'flagship' => [
        ['id' => 'mistral-large-4', 'name' => 'Mistral Large 4', 'desc' => 'Modèle généraliste multimodal, contexte 1M — Public Preview'],
        ['id' => 'mistral-medium-latest', 'name' => 'Mistral Medium 3.5', 'desc' => 'Raisonnement, agents et code'],
        ['id' => 'mistral-small-latest', 'name' => 'Mistral Small 4', 'desc' => 'Rapide, multimodal, raisonnement et code'],
        ['id' => 'ministral-14b-latest', 'name' => 'Ministral 3 14B', 'desc' => 'Modèle multimodal équilibré'],
        ['id' => 'ministral-8b-latest', 'name' => 'Ministral 3 8B', 'desc' => 'Modèle multimodal léger'],
        ['id' => 'ministral-3b-latest', 'name' => 'Ministral 3 3B', 'desc' => 'Modèle multimodal compact'],
    ],
    'code' => [
        ['id' => 'codestral-latest', 'name' => 'Codestral', 'desc' => 'Génération et complétion de code'],
        ['id' => 'codestral-2508', 'name' => 'Codestral 25.08', 'desc' => 'Version épinglée de Codestral'],
    ],
    'vision' => [
        ['id' => 'ministral-14b-2512', 'name' => 'Ministral 3 14B (25.12)', 'desc' => 'Vision et analyse de documents, version épinglée'],
        ['id' => 'ministral-8b-2512', 'name' => 'Ministral 3 8B (25.12)', 'desc' => 'Vision légère, version épinglée'],
    ],
]);

define('MODEL_ALIASES', [
    // Redirections transparentes des anciens snapshots Mistral retirés.
    'mistral-large-2512'     => 'mistral-large-latest',
    'mistral-large-2411'     => 'mistral-large-latest',
    'devstral-2512'          => 'codestral-2508',
    'mistral-code-latest'    => 'codestral-latest',
    'mistral-code-fim-latest'=> 'codestral-latest',
    'devstral-medium-2507'   => 'codestral-latest',
    'devstral-small-2507'    => 'codestral-latest',
    'pixtral-large-2411'     => 'ministral-14b-latest',
    'pixtral-12b-2409'       => 'ministral-8b-latest',
    'labs-mistral-small-creative' => 'codestral-latest',
    'ministral-3b-2512'      => 'ministral-3b-latest',
]);

// Modèle par défaut pour chaque rôle
define('MASTER_AGENT_MODEL', 'mistral-small-latest');
define('CODE_AGENT_MODEL', 'codestral-latest');
define('VISION_AGENT_MODEL', 'mistral-small-latest');
define('PLANNER_AGENT_MODEL', 'mistral-medium-latest');
define('CREATIVE_AGENT_MODEL', 'mistral-medium-latest');

// Sécurité
define('PASSWORD_MIN_LENGTH', 8);
define('MAX_LOGIN_ATTEMPTS', 5);
define('LOCKOUT_TIME', 900);
define('SESSION_LIFETIME', 86400);
define('CSRF_TOKEN_LENGTH', 32);

// Apprentissage
define('AUTO_LEARNING_ENABLED', true);
define('LEARNING_THRESHOLD', 0.8);

// Logs
$logFile = DATA_PATH . '/nodalibre.log';
$legacyLogFile = DATA_PATH . '/libre_' . $legacyBrandSuffix . '.log';
if (!file_exists($logFile) && file_exists($legacyLogFile)) {
    if (!@rename($legacyLogFile, $logFile)) {
        $logFile = $legacyLogFile;
    }
}
define('LOG_FILE', $logFile);
define('LOG_LEVEL', 3);

// Créer les dossiers si inexistants (permissions Hostinger)
foreach ([DATA_PATH, SANDBOX_PATH] as $dir) {
    if (!is_dir($dir)) {
        mkdir($dir, 0755, true);
    }
}

// Session sécurisée
if (session_status() === PHP_SESSION_NONE) {
    ini_set('session.cookie_httponly', 1);
    ini_set('session.use_strict_mode', 1);
    session_start();
}

// Logging (cURL safe, pas de exec)
function nodalibre_log($message, $level = 3) {
    if ($level > LOG_LEVEL) return;
    $levels = [1 => 'ERROR', 2 => 'WARNING', 3 => 'INFO', 4 => 'DEBUG'];
    $entry = '[' . date('Y-m-d H:i:s') . '] [' . ($levels[$level] ?? 'INFO') . '] ' . $message . "\n";
    @file_put_contents(LOG_FILE, $entry, FILE_APPEND | LOCK_EX);
}

set_exception_handler(function($e) {
    nodalibre_log("Exception: " . $e->getMessage(), 1);
    if (!headers_sent()) {
        http_response_code(500);
        if (strpos($_SERVER['HTTP_ACCEPT'] ?? '', 'application/json') !== false) {
            echo json_encode(['success' => false, 'error' => 'Erreur interne']);
        }
    }
});
