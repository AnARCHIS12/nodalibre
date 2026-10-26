## Nouveautés v1.2.0

### Catalogue Mistral
- **Mistral Large** — `mistral-large-latest`
- **Mistral Medium** — `mistral-medium-2604`
- **Mistral Small** — `mistral-small-latest`
- Les modèles apparaissent en tête de chaque catégorie dans le sélecteur.

### Administration : utilisateurs connectés
- Nouveau bloc en haut de la page `admin_keys.php` affichant le nombre d'utilisateurs actifs
- Liste des utilisateurs connectés (session active dans les 30 dernières minutes)
- Indicateur de présence : nom, adresse IP, heure de dernière activité
- Point vert animé (pulse) pour chaque utilisateur en ligne
- Badge "vous" pour distinguer l'administrateur connecté
- Auto-refresh de la page toutes les 60 secondes

### Points forts

- Interface Nodalibre rouge/noir avec historique utilisateur
- Image Docker Hub multi-architecture linux/amd64 et linux/arm64 : liberchat/nodalibre:latest
- Installateurs interactifs Linux/macOS et Windows
- Workspace GitHub avec OAuth, création de dépôts, édition de fichiers et commits multi-fichiers
- Mémoire automatique utilisateur et contexte workspace
- Recherche dans l'historique des conversations
- Export de conversations en Markdown et JSON
- **Trois modèles Mistral généralistes**
- **Tableau de bord admin : utilisateurs connectés en temps réel**
- Dictée vocale Voxtral
- Interface multilingue : français, anglais, espagnol, allemand et italien
- Licence MIT et README complet avec aperçu de l'interface

## Installation rapide

### Linux/macOS :
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
