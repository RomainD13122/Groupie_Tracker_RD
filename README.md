<<<<<<< HEAD
# PokéTracker

![PokéTracker Logo](static/img/logo.png)

PokéTracker est une application web développée en Go permettant d'explorer et de suivre les cartes Pokémon. Ce projet a été réalisé dans le cadre du projet "Groupie Tracker" pour démontrer l'utilisation d'une API REST et l'implémentation de fonctionnalités web essentielles.

## Fonctionnalités

- **Navigation de cartes** : Parcourez des milliers de cartes Pokémon
- **Recherche** : Recherchez des cartes par nom
- **Filtrage** : Filtrez les cartes par type, rareté et collection
- **Pagination** : Parcourez les résultats page par page
- **Favoris** : Ajoutez vos cartes préférées à une liste de favoris persistante
- **Détails des cartes** : Consultez les informations détaillées de chaque carte
- **Collections** : Explorez les différentes collections de cartes Pokémon

## Captures d'écran

![Page d'accueil](screenshots/home.png)
![Liste des cartes](screenshots/cards.png)
![Détails d'une carte](screenshots/card_detail.png)
![Favoris](screenshots/favorites.png)

## Technologies utilisées

- [Go](https://golang.org/) - Langage de programmation backend
- Bibliothèques standard Go uniquement :
  - `net/http` - Serveur web et requêtes HTTP
  - `html/template` - Gestion des templates HTML
  - `encoding/json` - Traitement des données JSON
- HTML/CSS - Interface utilisateur
- JavaScript vanille - Interactivité côté client
- [API TCGdex](https://api.tcgdex.net/) - API utilisée pour récupérer les données des cartes Pokémon

## Prérequis

- Go 1.20 ou supérieur

## Installation

1. Clonez ce dépôt :
   ```bash
   git clone https://github.com/votre-nom/poketracker.git
   cd poketracker
   ```

2. Exécutez l'application :
   ```bash
   go run main.go
   ```

3. Ouvrez votre navigateur et accédez à [http://localhost:8080](http://localhost:8080)

## Structure du projet

```
poketracker/
├── data/               # Stockage des données (favoris)
├── static/             # Fichiers statiques
│   └── css/            # Feuilles de style CSS
├── templates/          # Templates HTML
├── main.go             # Point d'entrée principal
├── go.mod              # Définition du module Go
└── README.md           # Ce fichier
```

## Routes implémentées

| Route | Description |
|-------|-------------|
| `/` | Page d'accueil |
| `/cards` | Liste des cartes avec filtres et pagination |
| `/card/{id}` | Détails d'une carte spécifique |
| `/sets` | Liste des collections |
| `/set/{id}` | Détails d'une collection et ses cartes |
| `/search?q={query}` | Recherche de cartes |
| `/favorites` | Liste des cartes favorites |
| `/api/favorite/add/{id}` | Ajouter une carte aux favoris |
| `/api/favorite/remove/{id}` | Retirer une carte des favoris |
| `/api/favorite/clear` | Vider la liste des favoris |
| `/about` | Page à propos avec informations sur le projet |

## API utilisée

Cette application utilise l'API TCGdex pour récupérer les informations sur les cartes Pokémon.

### Endpoints utilisés

| Endpoint | Description | Utilisation dans PokéTracker |
|----------|-------------|------------------------|
| `/v2/en/cards` | Récupération de la liste des cartes | Page des cartes, recherche et filtrage |
| `/v2/en/cards/{id}` | Récupération des détails d'une carte | Page de détail d'une carte |
| `/v2/en/sets` | Récupération de la liste des collections | Page des collections et options de filtrage |
| `/v2/en/sets/{id}` | Récupération des détails d'une collection | Page de détail d'une collection |
| `/v2/en/types` | Récupération de la liste des types de cartes | Options de filtrage par type |
| `/v2/en/rarities` | Récupération de la liste des raretés de cartes | Options de filtrage par rareté |

## Gestion des erreurs

L'application inclut une gestion robuste des erreurs :
- Gestion des cas où l'API est indisponible avec mécanisme de retry
- Affichage de messages d'erreur explicites à l'utilisateur
- Utilisation de valeurs par défaut et solutions de secours en cas d'erreur

## Développement et maintenance

Pour contribuer au projet :

1. Forker le dépôt
2. Créer une branche pour votre fonctionnalité (`git checkout -b feature/AmazingFeature`)
3. Commiter vos changements (`git commit -m 'Add some AmazingFeature'`)
4. Pousser la branche (`git push origin feature/AmazingFeature`)
5. Ouvrir une Pull Request

## Points d'amélioration futurs

- Authentification des utilisateurs
- Statistiques avancées sur les collections et les cartes
- Système de suivi de prix et de valeur des cartes
- Support multilingue
- Mode dark/light

## Auteur

- **Romain-Daniel** - [GitHub](https://github.com/RomainD13122/Groupie_Tracker_RD.git)

## Remerciements

- [TCGdex](https://api.tcgdex.net/) pour leur API gratuite et complète
- Le projet Pokémon pour l'univers et les cartes
=======
# Groupie_Tracker_RD
Groupie Tracker date limite 10/03/2025
>>>>>>> 475af3ab1473ca17767de9d423d30ac69783e030
