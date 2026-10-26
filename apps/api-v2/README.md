> [!WARNING]
> Document historique : projet abandonné, plus maintenu. La v1.4.1 est la
> dernière version. Les travaux et migrations décrits ci-dessous sont annulés ;
> ce document est conservé uniquement pour une éventuelle reprise indépendante.

# Nodalibre V2

Ce répertoire contient la réécriture Go multi-provider. Il est volontairement
isolé de la V1 PHP tant que la compatibilité fonctionnelle et la migration
SQLite ne sont pas validées.

État actuel :

- contrat de fournisseur indépendant ;
- registre de fournisseurs ;
- découverte et streaming Mistral ;
- découverte et streaming ChatGPT Responses avec un jeton OAuth déjà obtenu ;
- tests unitaires du registre et des contraintes de fin de stream ChatGPT.

Prochain lot : stockage chiffré des comptes fournisseurs, migrations SQLite,
authentification « Sign in with ChatGPT », puis routes de compatibilité V1.

```bash
cd apps/api-v2
go test ./...
```
