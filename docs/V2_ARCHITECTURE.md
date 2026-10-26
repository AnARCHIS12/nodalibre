# Nodalibre V2 — architecture multi-provider

## Décision technique

La V2 sera écrite en Go et livrée sous forme d'un binaire unique avec les
ressources web embarquées. SQLite reste la base par défaut afin de conserver
les comptes, conversations, messages, mémoires et workspaces de la V1.

Go est retenu pour :

- une faible consommation mémoire et un démarrage rapide ;
- une image Docker minimale sans serveur Apache ni runtime PHP ;
- une bonne bibliothèque HTTP, OAuth 2/OIDC et SSE ;
- des déploiements simples en `amd64` et `arm64` ;
- une maintenance plus accessible que Rust pour cette application web.

## Principes

1. Aucun fournisseur n'est spécial dans le domaine métier.
2. Les noms et identifiants réels des modèles sont affichés.
3. Les modèles disponibles sont découverts auprès du compte quand le
   fournisseur le permet ; ils ne sont pas figés dans l'interface.
4. Une conversation conserve le fournisseur et le modèle utilisés pour chaque
   message, mais peut changer de fournisseur au message suivant.
5. Les capacités sont déclarées séparément : texte, vision, fichiers,
   recherche web, image, transcription et synthèse vocale.
6. Une capacité absente chez un fournisseur est désactivée dans l'interface au
   lieu d'être simulée ou redirigée silencieusement.

## Fournisseurs prévus

### Mistral

- Authentification par clé API personnelle ou pool administrateur.
- Chat Completions pour le texte et la vision.
- Conversations pour la recherche web et les outils intégrés.
- OCR, transcription Voxtral, TTS et génération d'images conservés.
- Catalogue récupéré par API puis filtré selon les capacités utiles.

### ChatGPT

- Authentification officielle « Sign in with ChatGPT ».
- Autorisation facultative `chatgpt.tokens.use.direct` pour utiliser le forfait
  ChatGPT de l'utilisateur sans clé API.
- Catalogue récupéré depuis `GET https://api.openai.com/v1/models` avec le
  jeton OAuth du compte sélectionné.
- Inférence via `POST https://api.openai.com/v1/responses`, avec `store: false`
  et `stream: true`.
- Aucun accès aux conversations ou à la mémoire ChatGPT de l'utilisateur.

L'authentification initiale utilise un callback local sur `127.0.0.1`. Pour une
instance installée sur un VPS ou un serveur domestique, la connexion est
effectuée depuis un petit client local puis les identifiants protégés sont
transférés vers l'instance. Un téléphone peut ensuite utiliser le compte déjà
connecté depuis l'interface web.

### Google Gemini

- OAuth 2 officiel vers la Gemini API, avec un projet Google Cloud et un client
  OAuth configurés par l'administrateur.
- Mode clé API également disponible comme solution de repli.
- L'abonnement Google AI Pro/Ultra ne couvre pas les requêtes d'une application
  tierce : OAuth évite la saisie d'une clé par l'utilisateur, mais ne transforme
  pas l'abonnement grand public en quota API.
- Aucun appel aux endpoints internes d'Antigravity ou de Gemini Web.

### Local et compatible OpenAI

- Fournisseur OpenAI-compatible paramétrable pour Ollama, llama.cpp, vLLM ou
  un proxy explicitement choisi par l'administrateur.
- URL locale autorisée seulement après validation afin d'éviter les requêtes
  serveur arbitraires (SSRF).

## Contrat interne

Le coeur utilise des interfaces petites et indépendantes :

```go
type ChatProvider interface {
    ID() string
    ListModels(ctx context.Context, account Account) ([]Model, error)
    Stream(ctx context.Context, req ChatRequest, sink EventSink) error
}

type CapabilityProvider interface {
    Capabilities(model Model) Capabilities
}
```

Les fonctions OCR, transcription, synthèse vocale et image utilisent des
interfaces optionnelles distinctes. Un fournisseur de chat n'a donc pas besoin
d'implémenter artificiellement toutes les fonctions de Mistral.

## Évolution SQLite

Les tables V1 sont conservées. Les migrations sont uniquement additives au
début :

- `provider_accounts` : compte fournisseur lié à un utilisateur ;
- `provider_credentials` : secrets chiffrés, séparés des métadonnées ;
- `provider_models` : cache court du catalogue découvert ;
- `provider_usage` : consommation normalisée et détails propres au fournisseur ;
- ajout de `provider_id` et `provider_account_id` aux conversations/messages.

Les jetons OAuth, refresh tokens, clés Mistral et jetons GitHub seront chiffrés
avec une clé maîtresse extérieure à SQLite. Les anciennes colonnes restent
lisibles pendant la transition, puis sont vidées après migration vérifiée.

## Compatibilité HTTP

La V2 expose une API stable sous `/api/v2`. Pendant la transition, des routes
de compatibilité reprennent les réponses JSON attendues par l'interface V1 :

- `/chat.php` ;
- `/conversations.php` ;
- `/transcribe.php` ;
- `/speak.php` ;
- `/ocr.php` ;
- `/image_generate.php` ;
- `/workspace_api.php`.

Cela permet de migrer le backend puis l'interface sans casser les clients
existants. La version finale ne contient plus PHP ni Apache.

## Étapes de migration

1. Corriger et stabiliser la V1 Mistral.
2. Créer le coeur Go, le registre de fournisseurs et les migrations SQLite.
3. Porter l'authentification locale, les sessions, conversations et messages.
4. Porter Mistral à comportement équivalent et valider les routes de
   compatibilité.
5. Ajouter ChatGPT OAuth + Responses API et le sélecteur de comptes/modèles.
6. Ajouter Google OAuth/API, en affichant clairement le mode de facturation.
7. Porter fichiers, OCR, voix, images, mémoire, workspace et GitHub.
8. Extraire l'interface PHP en ressources web statiques embarquées.
9. Tester la migration d'une copie de la base V1, puis proposer le basculement.
10. Retirer le conteneur PHP seulement après validation fonctionnelle complète.

## Critères de sortie V2

- aucune perte sur une copie réelle de la base V1 ;
- toutes les fonctions documentées de la V1 couvertes par des tests ;
- changement de fournisseur sans recréer la conversation ;
- révocation et renouvellement OAuth testés ;
- secrets absents des logs et chiffrés au repos ;
- images Docker `amd64` et `arm64` ;
- procédure de retour à la V1 documentée tant que la migration n'est pas
  déclarée définitive.
