## Nouveautés v1.3.0

### Catalogue de modèles Mistral
- **Codestral** (`codestral-latest`) pour la génération de code.
- **Ministral 14B** (`ministral-14b-latest`) pour le texte, la vision et les documents.
- **Ministral 8B** (`ministral-8b-latest`) pour les usages rapides et multimodaux.
- **Ministral 3B** (`ministral-3b-latest`) pour les commandes courtes et légères.
- Les modèles sont affichés sous leurs noms et identifiants fournisseur réels.

### Fallback Vision OCR intelligent
- Si le endpoint OCR dédié (`/v1/ocr`) est limité sur un compte gratuit (`HTTP 429`), l'application bascule automatiquement sur la vision multimodale (`ministral-14b-latest`) pour transcrire les documents et images sans interruption.

### Recherche Web fiabilisée
- Routage automatique des requêtes avec recherche web vers les modèles compatibles avec les connecteurs Mistral (`mistral-small-latest`).

### Nodalibre Coder fiabilisé (Workspace)
- L'agent de code du workspace utilise désormais directement le modèle dédié `codestral-2508` pour des modifications et des commits multi-fichiers instantanés.

### Installateur Windows interactif amélioré (`install.ps1`)
- Détection et installation automatisée de WSL 2 sans blocage.
- Timeout sur les commandes natives pour éviter les gels du terminal.
- Polling du daemon Docker avec journalisation de l'état.
- Anonymisation automatique des chemins utilisateur à l'écran (`~` au lieu de `C:\Users\nom`).
- Gestion de la fermeture de fenêtre PowerShell pour afficher les instructions finales.

### Gestion transparente des erreurs API
- Affichage explicite des messages retournés par l'API au lieu du message opaque générique.

---

## Installation rapide

### Linux / macOS :
```bash
curl -fsSL https://raw.githubusercontent.com/AnARCHIS12/Nodalibre/main/install.sh | sh
```

### Windows PowerShell :
```powershell
irm https://raw.githubusercontent.com/AnARCHIS12/Nodalibre/main/install.ps1 | iex
```

### Docker :
```bash
docker run -d \
  --name nodalibre \
  -p 8173:80 \
  -v "$PWD/data:/var/www/html/data" \
  -v "$PWD/sandbox:/var/www/html/sandbox" \
  liberchat/nodalibre:latest
```
