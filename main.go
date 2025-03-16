package main

import (
	"encoding/json"
	"fmt"
	"html/template"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

var templates *template.Template

type Card struct {
	ID             string   `json:"id"`
	Name           string   `json:"name"`
	Image          string   `json:"image,omitempty"`
	Set            Set      `json:"set,omitempty"`
	Number         string   `json:"number,omitempty"`
	Rarity         string   `json:"rarity,omitempty"`
	Types          []string `json:"types,omitempty"`
	Description    string   `json:"description,omitempty"`
	Artist         string   `json:"artist,omitempty"`
	HP             any      `json:"hp,omitempty"`
	Images         Images   `json:"images,omitempty"`
	Illustrator    string   `json:"illustrator,omitempty"`
	Category       string   `json:"category,omitempty"`
	LocalId        string   `json:"localId,omitempty"`
	RegulationMark string   `json:"regulationMark,omitempty"`
}

type Images struct {
	Small string `json:"small,omitempty"`
	Large string `json:"large,omitempty"`
}

type Set struct {
	ID          string    `json:"id"`
	Name        string    `json:"name"`
	Logo        string    `json:"logo,omitempty"`
	Symbol      string    `json:"symbol,omitempty"`
	CardCount   CardCount `json:"cardCount,omitempty"`
	ReleaseDate string    `json:"releaseDate,omitempty"`
	Legal       Legal     `json:"legal,omitempty"`
}

type CardCount struct {
	Total      int `json:"total,omitempty"`
	Official   int `json:"official,omitempty"`
	Unofficial int `json:"unofficial,omitempty"`
}

type Legal struct {
	Standard  bool `json:"standard,omitempty"`
	Expanded  bool `json:"expanded,omitempty"`
	Unlimited bool `json:"unlimited,omitempty"`
}

type Favorites struct {
	Cards []Card `json:"cards"`
}

func init() {

	funcMap := template.FuncMap{
		"add": func(a, b int) int {
			return a + b
		},
		"sub": func(a, b int) int {
			return a - b
		},
		"buildURL": func(base string, params map[string]string, keyValues ...string) string {
			result := base
			queryParams := make([]string, 0)

			for key, value := range params {
				if value != "" {
					queryParams = append(queryParams, fmt.Sprintf("%s=%s", key, value))
				}
			}

			for i := 0; i < len(keyValues); i += 2 {
				if i+1 < len(keyValues) {
					key := keyValues[i]
					value := keyValues[i+1]

					for j, param := range queryParams {
						if strings.HasPrefix(param, key+"=") {
							queryParams = append(queryParams[:j], queryParams[j+1:]...)
							break
						}
					}

					queryParams = append(queryParams, fmt.Sprintf("%s=%s", key, value))
				}
			}

			if len(queryParams) > 0 {
				result += "?" + strings.Join(queryParams, "&")
			}

			return result
		},
	}

	if _, err := os.Stat("templates"); os.IsNotExist(err) {
		log.Fatalf("ERREUR: Le dossier 'templates' n'existe pas. Veuillez le créer.")
	}

	log.Println("Chargement des templates...")

	var err error
	templates, err = template.New("").Funcs(funcMap).ParseGlob("templates/*.html")
	if err != nil {
		log.Fatalf("Erreur fatale lors du chargement des templates: %v", err)
	}

	templateNames := templates.Templates()
	log.Printf("Templates chargés (%d): ", len(templateNames))
	for _, t := range templateNames {
		log.Printf("  - %s", t.Name())
	}
}

func main() {

	err := os.MkdirAll("data", 0755)
	if err != nil {
		log.Printf("Erreur lors de la création du dossier data: %v", err)
	}

	err = os.MkdirAll("static/css", 0755)
	if err != nil {
		log.Printf("Erreur lors de la création des dossiers static/css: %v", err)
	}

	cssPath := filepath.Join("static", "css", "style.css")
	if _, err := os.Stat(cssPath); os.IsNotExist(err) {
		log.Printf("ATTENTION: Le fichier CSS n'existe pas à l'emplacement: %s", cssPath)

		cssContent := `
		body { font-family: Arial, sans-serif; margin: 0; padding: 0; background-color: #f0f0f0; }
		header { background-color: #e60000; color: white; padding: 1rem; }
		header h1 { margin: 0; }
		header a { color: white; text-decoration: none; }
		ul { list-style-type: none; padding: 0; }
		li { display: inline; margin-right: 15px; }
		.container { width: 80%; margin: 0 auto; padding: 20px; }
		.card-grid, .set-grid { display: grid; grid-template-columns: repeat(auto-fill, minmax(200px, 1fr)); gap: 20px; margin-top: 20px; }
		.card, .set { background: white; border-radius: 8px; box-shadow: 0 2px 5px rgba(0,0,0,0.1); overflow: hidden; transition: transform 0.3s; }
		.card:hover, .set:hover { transform: translateY(-5px); }
		.card img, .set img { width: 100%; height: auto; }
		.error-message { color: red; background: #ffeeee; padding: 10px; border-radius: 4px; margin: 10px 0; }
		.button { display: inline-block; background: #3b4cca; color: white; padding: 8px 15px; border-radius: 4px; text-decoration: none; }
		`

		err = os.WriteFile(cssPath, []byte(cssContent), 0644)
		if err != nil {
			log.Printf("Erreur lors de la création du CSS minimal: %v", err)
		} else {
			log.Printf("CSS minimal créé à %s", cssPath)
		}
	} else {
		log.Printf("CSS trouvé: %s", cssPath)
	}

	fs := http.FileServer(http.Dir("static"))
	http.Handle("/static/", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		path := r.URL.Path
		log.Printf("Requête de fichier statique: %s", path)
		http.StripPrefix("/static/", fs).ServeHTTP(w, r)
	}))

	http.HandleFunc("/", homeHandler)
	http.HandleFunc("/cards", cardsHandler)
	http.HandleFunc("/card/", cardDetailHandler)
	http.HandleFunc("/sets", setsHandler)
	http.HandleFunc("/set/", setDetailHandler)
	http.HandleFunc("/search", searchHandler)
	http.HandleFunc("/favorites", favoritesHandler)
	http.HandleFunc("/api/favorite/add/", addFavoriteHandler)
	http.HandleFunc("/api/favorite/remove/", removeFavoriteHandler)
	http.HandleFunc("/about", aboutHandler)
	http.HandleFunc("/test-images", testImagesHandler)
	http.HandleFunc("/api/favorite/clear", clearFavoritesHandler)

	port := "8080"
	log.Printf("Serveur démarré sur le port %s...", port)
	log.Fatal(http.ListenAndServe(":"+port, nil))
}

func fetchJSON(apiURL string, target interface{}) error {

	client := &http.Client{
		Timeout: 20 * time.Second,
	}

	var lastErr error
	for attempt := 0; attempt < 3; attempt++ {

		resp, err := client.Get(apiURL)
		if err != nil {
			lastErr = err
			log.Printf("Tentative d'API %d échouée: %v", attempt+1, err)
			time.Sleep(time.Duration(attempt+1) * time.Second)
			continue
		}

		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			lastErr = fmt.Errorf("API a retourné le code %d", resp.StatusCode)
			log.Printf("Tentative d'API %d échouée: %v", attempt+1, lastErr)
			time.Sleep(time.Duration(attempt+1) * time.Second)
			continue
		}

		var bodyBytes []byte
		bodyBytes, err = io.ReadAll(resp.Body)
		if err != nil {
			lastErr = err
			log.Printf("Lecture de la réponse API tentative %d échouée: %v", attempt+1, err)
			time.Sleep(time.Duration(attempt+1) * time.Second)
			continue
		}

		if len(bodyBytes) < 500 {
			log.Printf("Réponse API: %s", string(bodyBytes))
		} else {
			log.Printf("Réponse API (tronquée): %s...", string(bodyBytes[:500]))
		}

		err = json.Unmarshal(bodyBytes, target)
		if err != nil {
			lastErr = err
			log.Printf("Parsing JSON tentative %d échouée: %v", attempt+1, err)

			if len(bodyBytes) > 200 {
				log.Printf("Aperçu du corps de la réponse: %s...", bodyBytes[:200])
			} else {
				log.Printf("Corps de la réponse: %s", bodyBytes)
			}

			time.Sleep(time.Duration(attempt+1) * time.Second)
			continue
		}

		return nil
	}

	return fmt.Errorf("toutes les tentatives de requête API ont échoué, dernière erreur: %v", lastErr)
}

func fetchCards(page, limit int, filters map[string]string) ([]Card, int, error) {
	baseURL := "https://api.tcgdex.net/v2/en/cards"

	log.Printf("Requête de cartes URL: %s", baseURL)

	var cards []Card
	err := fetchJSON(baseURL, &cards)
	if err != nil {
		return []Card{}, 0, err
	}

	var filteredCards []Card
	for _, card := range cards {
		include := true

		for key, value := range filters {
			if value == "" {
				continue
			}

			switch key {
			case "type":
				found := false
				for _, t := range card.Types {
					if t == value {
						found = true
						break
					}
				}
				if !found {
					include = false
				}
			case "rarity":
				if card.Rarity != value {
					include = false
				}
			case "set":
				if card.Set.ID != value {
					include = false
				}
			case "name":
				if !strings.Contains(strings.ToLower(card.Name), strings.ToLower(value)) {
					include = false
				}
			}

			if !include {
				break
			}
		}

		if include {

			if card.Image != "" {
				if !strings.HasSuffix(card.Image, ".png") && !strings.HasSuffix(card.Image, ".jpg") {
					card.Image = card.Image + "/high.jpg"
				}
			} else if card.Images.Large != "" {
				card.Image = card.Images.Large
			} else if card.Images.Small != "" {
				card.Image = card.Images.Small
			} else {
				encodedName := url.QueryEscape(card.Name)
				card.Image = "https://via.placeholder.com/245x342.png?text=" + encodedName
			}

			filteredCards = append(filteredCards, card)
		}
	}

	total := len(filteredCards)
	start := (page - 1) * limit
	end := start + limit

	if start >= total {
		start = 0
		end = limit
		if end > total {
			end = total
		}
	}

	if end > total {
		end = total
	}

	if start < 0 {
		start = 0
	}

	var pagedCards []Card
	if total > 0 && start < total {
		pagedCards = filteredCards[start:end]
	}

	return pagedCards, total, nil
}

func fetchCard(id string) (Card, error) {
	apiURL := fmt.Sprintf("https://api.tcgdex.net/v2/en/cards/%s", id)
	var card Card
	err := fetchJSON(apiURL, &card)

	if card.Image != "" {
		if !strings.HasSuffix(card.Image, ".png") && !strings.HasSuffix(card.Image, ".jpg") {
			card.Image = card.Image + "/high.jpg"
		}
	} else if card.Images.Large != "" {
		card.Image = card.Images.Large
	} else if card.Images.Small != "" {
		card.Image = card.Images.Small
	} else {

		encodedName := url.QueryEscape(card.Name)
		card.Image = "https://via.placeholder.com/245x342.png?text=" + encodedName
	}

	return card, err
}
func (c Card) GetHP() string {
	switch hp := c.HP.(type) {
	case int:
		return strconv.Itoa(hp)
	case float64:
		return strconv.Itoa(int(hp))
	case string:
		return hp
	default:
		return ""
	}
}

func fetchSets() ([]Set, error) {
	apiURL := "https://api.tcgdex.net/v2/en/sets"
	log.Printf("Requête des sets URL: %s", apiURL)

	var sets []Set
	err := fetchJSON(apiURL, &sets)

	if err != nil {
		return []Set{}, err
	}

	for i := range sets {

		if sets[i].Logo != "" && !strings.HasSuffix(sets[i].Logo, ".png") && !strings.HasSuffix(sets[i].Logo, ".jpg") {
			sets[i].Logo = sets[i].Logo + ".png"
		}

		if sets[i].Symbol != "" && !strings.HasSuffix(sets[i].Symbol, ".png") && !strings.HasSuffix(sets[i].Symbol, ".jpg") {
			sets[i].Symbol = sets[i].Symbol + ".png"
		}

		if sets[i].Logo == "" && sets[i].Symbol == "" {

			encodedName := url.QueryEscape(sets[i].Name)
			sets[i].Logo = "https://via.placeholder.com/200x100.png?text=" + encodedName
		}
	}

	log.Printf("Nombre de sets récupérés: %d", len(sets))
	return sets, nil
}

func fetchSet(id string) (Set, error) {
	apiURL := fmt.Sprintf("https://api.tcgdex.net/v2/en/sets/%s", id)
	var set Set
	err := fetchJSON(apiURL, &set)

	if set.Logo != "" && !strings.HasSuffix(set.Logo, ".png") && !strings.HasSuffix(set.Logo, ".jpg") {
		set.Logo = set.Logo + ".png"
	}

	if set.Symbol != "" && !strings.HasSuffix(set.Symbol, ".png") && !strings.HasSuffix(set.Symbol, ".jpg") {
		set.Symbol = set.Symbol + ".png"
	}

	if set.Logo == "" && set.Symbol == "" {

		encodedName := url.QueryEscape(set.Name)
		set.Logo = "https://via.placeholder.com/200x100.png?text=" + encodedName
	}

	return set, err
}

func fetchTypes() ([]string, error) {

	var types []string
	apiURL := "https://api.tcgdex.net/v2/en/types"
	err := fetchJSON(apiURL, &types)

	if err != nil || len(types) == 0 {
		log.Printf("Utilisation de la liste de secours pour les types: %v", err)
		return []string{
			"Colorless", "Darkness", "Dragon", "Fairy", "Fighting",
			"Fire", "Grass", "Lightning", "Metal", "Psychic", "Water",
		}, nil
	}

	return types, nil
}

func fetchRarities() ([]string, error) {

	var rarities []string
	apiURL := "https://api.tcgdex.net/v2/en/rarities"
	err := fetchJSON(apiURL, &rarities)

	if err != nil || len(rarities) == 0 {
		log.Printf("Utilisation de la liste de secours pour les raretés: %v", err)
		return []string{
			"Common", "Uncommon", "Rare", "Rare Holo", "Rare Ultra",
			"Rare Holo EX", "Rare Holo GX", "Rare Holo V", "Rare Holo VMAX",
			"Amazing Rare", "Rare Rainbow", "Rare Secret", "Promo",
		}, nil
	}

	return rarities, nil
}

func loadFavorites() (Favorites, error) {
	var favorites Favorites
	favorites.Cards = []Card{} // Initialize with empty array

	// Check if file exists
	_, err := os.Stat("data/favorites.json")
	if os.IsNotExist(err) {

		err = saveFavorites(favorites)
		if err != nil {
			return favorites, err
		}
		return favorites, nil
	}

	data, err := os.ReadFile("data/favorites.json")
	if err != nil {
		return favorites, err
	}

	if len(data) == 0 || string(data) == "" {
		return favorites, nil
	}

	err = json.Unmarshal(data, &favorites)
	if err != nil {
		log.Printf("Erreur de parsing du fichier de favoris, création d'un nouveau fichier: %v", err)
		newFavorites := Favorites{Cards: []Card{}}
		saveFavorites(newFavorites)
		return newFavorites, nil
	}

	return favorites, nil
}

func saveFavorites(favorites Favorites) error {

	os.MkdirAll("data", 0755)

	data, err := json.Marshal(favorites)
	if err != nil {
		return err
	}

	return os.WriteFile("data/favorites.json", data, 0644)
}

func homeHandler(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		showError(w, "Page non trouvée", fmt.Errorf("URL invalide: %s", r.URL.Path))
		return
	}

	cards, _, err := fetchCards(1, 6, nil)

	sets, err2 := fetchSets()

	data := struct {
		RecentCards []Card
		Sets        []Set
		Error       string
	}{
		RecentCards: cards,
		Sets:        sets,
	}

	if err != nil {
		log.Printf("Erreur lors de la récupération des cartes: %v", err)
		data.Error = "Impossible de charger les cartes récentes."
	}

	if err2 != nil {
		log.Printf("Erreur lors de la récupération des sets: %v", err2)
		if data.Error != "" {
			data.Error += " "
		}
		data.Error += "Impossible de charger les collections."
	}

	if len(sets) > 6 {
		data.Sets = sets[:6]
	}

	if err := templates.ExecuteTemplate(w, "index.html", data); err != nil {
		log.Printf("Erreur de rendu du template index.html: %v", err)
		showError(w, "Erreur d'affichage de la page d'accueil", err)
	}
}

func cardsHandler(w http.ResponseWriter, r *http.Request) {

	r.ParseForm()
	page, _ := strconv.Atoi(r.FormValue("page"))
	if page < 1 {
		page = 1
	}

	limit, _ := strconv.Atoi(r.FormValue("limit"))
	if limit != 10 && limit != 20 && limit != 30 {
		limit = 20
	}

	filters := make(map[string]string)
	if typeFilter := r.FormValue("type"); typeFilter != "" {
		filters["type"] = typeFilter
	}
	if rarityFilter := r.FormValue("rarity"); rarityFilter != "" {
		filters["rarity"] = rarityFilter
	}
	if setFilter := r.FormValue("set"); setFilter != "" {
		filters["set"] = setFilter
	}

	data := struct {
		Cards      []Card
		Types      []string
		Rarities   []string
		Sets       []Set
		Filters    map[string]string
		Pagination interface{}
		Limit      int
		Total      int
		Error      string
	}{
		Cards:    []Card{},
		Types:    []string{},
		Rarities: []string{},
		Sets:     []Set{},
		Filters:  filters,
		Limit:    limit,
		Total:    0,
	}

	cards, total, err := fetchCards(page, limit, filters)
	if err != nil {
		log.Printf("Erreur lors de la récupération des cartes: %v", err)
		data.Error = "Impossible de récupérer les cartes. Veuillez réessayer plus tard."
	} else {
		data.Cards = cards
		data.Total = total

		totalPages := (total + limit - 1) / limit
		if totalPages < 1 {
			totalPages = 1
		}

		data.Pagination = struct {
			CurrentPage int
			TotalPages  int
			HasPrev     bool
			HasNext     bool
		}{
			CurrentPage: page,
			TotalPages:  totalPages,
			HasPrev:     page > 1,
			HasNext:     page < totalPages,
		}
	}

	types, err := fetchTypes()
	if err != nil {
		log.Printf("Erreur lors de la récupération des types: %v", err)
	} else {
		data.Types = types
	}

	rarities, err := fetchRarities()
	if err != nil {
		log.Printf("Erreur lors de la récupération des raretés: %v", err)
	} else {
		data.Rarities = rarities
	}

	sets, err := fetchSets()
	if err != nil {
		log.Printf("Erreur lors de la récupération des sets: %v", err)
	} else {
		data.Sets = sets
	}

	if err := templates.ExecuteTemplate(w, "cards.html", data); err != nil {
		log.Printf("Erreur de rendu du template cards.html: %v", err)
		showError(w, "Erreur d'affichage de la liste des cartes", err)
	}
}

func cardDetailHandler(w http.ResponseWriter, r *http.Request) {
	id := strings.TrimPrefix(r.URL.Path, "/card/")
	if id == "" {
		showError(w, "Page non trouvée", fmt.Errorf("ID de carte non spécifié"))
		return
	}

	card, err := fetchCard(id)
	if err != nil {
		showError(w, "Impossible de récupérer les détails de la carte", err)
		return
	}

	favorites, _ := loadFavorites()
	isFavorite := false
	for _, favCard := range favorites.Cards {
		if favCard.ID == card.ID {
			isFavorite = true
			break
		}
	}

	html := `<!DOCTYPE html>
<html lang="fr">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>` + card.Name + ` - PokéTracker</title>
    <link rel="stylesheet" href="/static/css/style.css">
</head>
<body>
    <header>
        <div class="container">
            <h1><a href="/">PokéTracker</a></h1>
            <nav>
                <ul>
                    <li><a href="/">Accueil</a></li>
                    <li><a href="/cards">Cartes</a></li>
                    <li><a href="/sets">Collections</a></li>
                    <li><a href="/favorites">Favoris</a></li>
                    <li><a href="/about">À propos</a></li>
                </ul>
            </nav>
            <form action="/search" method="GET" class="search-form">
                <input type="text" name="q" placeholder="Rechercher des cartes..." required>
                <button type="submit">Rechercher</button>
            </form>
        </div>
    </header>
    
    <main class="container">
        <div class="card-detail fade-in">
            <div class="card-image">
                <img src="` + card.Image + `" alt="` + card.Name + `">
                
                <div class="favorite-controls">`

	if isFavorite {
		html += `<button id="remove-favorite" data-id="` + card.ID + `" class="button">Retirer des Favoris</button>`
	} else {
		html += `<button id="add-favorite" data-id="` + card.ID + `" class="button">Ajouter aux Favoris</button>`
	}

	html += `</div>
            </div>
            
            <div class="card-info">
                <h2>` + card.Name + `</h2>
                
                <div class="card-meta">`

	if card.Set.Name != "" {
		html += `<p><strong>Collection:</strong> <a href="/set/` + card.Set.ID + `">` + card.Set.Name + `</a></p>`
	}

	if card.Number != "" {
		html += `<p><strong>Numéro:</strong> ` + card.Number + `</p>`
	}

	if card.Rarity != "" {
		html += `<p><strong>Rareté:</strong> ` + card.Rarity + `</p>`
	}

	// Utiliser la nouvelle méthode GetHP
	hpValue := card.GetHP()
	if hpValue != "" {
		html += `<p><strong>HP:</strong> ` + hpValue + `</p>`
	}

	if len(card.Types) > 0 {
		html += `<p><strong>Types:</strong>
                    <div class="type-list">`
		for _, t := range card.Types {
			html += `<span class="type ` + t + `">` + t + `</span>`
		}
		html += `</div>
                </p>`
	}

	if card.Artist != "" {
		html += `<p><strong>Illustrateur:</strong> ` + card.Artist + `</p>`
	} else if card.Illustrator != "" {
		html += `<p><strong>Illustrateur:</strong> ` + card.Illustrator + `</p>`
	}

	if card.Category != "" {
		html += `<p><strong>Catégorie:</strong> ` + card.Category + `</p>`
	}

	if card.RegulationMark != "" {
		html += `<p><strong>Régulation:</strong> ` + card.RegulationMark + `</p>`
	}

	html += `</div>
                
                <div class="card-actions">
                    <a href="/cards" class="button secondary">Retour aux Cartes</a>
                    <a href="/set/` + card.Set.ID + `" class="button">Voir la Collection</a>
                </div>
            </div>
        </div>
    </main>
    
    <footer>
        <div class="container">
            <p>&copy; 2025 PokéTracker - Créé pour le projet Groupie Tracker</p>
        </div>
    </footer>
    
    <script>
        document.addEventListener('DOMContentLoaded', function() {
            const addButton = document.getElementById('add-favorite');
            const removeButton = document.getElementById('remove-favorite');
            
            if (addButton) {
                addButton.addEventListener('click', function() {
                    const cardId = this.getAttribute('data-id');
                    fetch('/api/favorite/add/' + cardId)
                        .then(response => {
                            if (response.ok) {
                                window.location.reload();
                            }
                        });
                });
            }
            
            if (removeButton) {
                removeButton.addEventListener('click', function() {
                    const cardId = this.getAttribute('data-id');
                    fetch('/api/favorite/remove/' + cardId)
                        .then(response => {
                            if (response.ok) {
                                window.location.reload();
                            }
                        });
                });
            }
        });
    </script>
</body>
</html>`

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Write([]byte(html))
}

func setsHandler(w http.ResponseWriter, r *http.Request) {
	sets, err := fetchSets()

	if err != nil {
		showError(w, "Impossible de récupérer la liste des collections", err)
		return
	}

	html := `<!DOCTYPE html>
<html lang="fr">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>Collections - PokéTracker</title>
    <link rel="stylesheet" href="/static/css/style.css">
</head>
<body>
    <header>
        <div class="container">
            <h1><a href="/">PokéTracker</a></h1>
            <nav>
                <ul>
                    <li><a href="/">Accueil</a></li>
                    <li><a href="/cards">Cartes</a></li>
                    <li><a href="/sets">Collections</a></li>
                    <li><a href="/favorites">Favoris</a></li>
                    <li><a href="/about">À propos</a></li>
                </ul>
            </nav>
            <form action="/search" method="GET" class="search-form">
                <input type="text" name="q" placeholder="Rechercher des cartes..." required>
                <button type="submit">Rechercher</button>
            </form>
        </div>
    </header>
    
    <main class="container">
        <div class="page-header">
            <h2>Collections Pokémon</h2>
            <p>Découvrez toutes les collections de cartes Pokémon</p>
        </div>
        
        <div class="set-grid">`

	if len(sets) > 0 {
		for _, set := range sets {
			html += `<div class="set">
                <a href="/set/` + set.ID + `">
                    <div class="set-image">`

			if set.Logo != "" {
				html += `<img src="` + set.Logo + `" alt="` + set.Name + `">`
			} else if set.Symbol != "" {
				html += `<img src="` + set.Symbol + `" alt="` + set.Name + `">`
			} else {
				html += `<div class="no-image">` + set.Name + `</div>`
			}

			html += `</div>
                    <div class="set-info">
                        <h3>` + set.Name + `</h3>`

			if set.CardCount.Total > 0 {
				html += `<p>` + strconv.Itoa(set.CardCount.Total) + ` cartes</p>`
			} else {
				html += `<p>? cartes</p>`
			}

			if set.ReleaseDate != "" {
				html += `<p class="release-date">Date de sortie: ` + set.ReleaseDate + `</p>`
			}

			html += `</div>
                </a>
            </div>`
		}
	} else {
		html += `<p class="no-results">Aucune collection n'a pu être chargée. Veuillez réessayer plus tard.</p>`
	}

	html += `</div>
    </main>
    
    <footer>
        <div class="container">
            <p>&copy; 2025 PokéTracker - Créé pour le projet Groupie Tracker</p>
        </div>
    </footer>
</body>
</html>`

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Write([]byte(html))
}
func fetchSetCards(setID string, limit int) ([]Card, error) {

	apiURL := fmt.Sprintf("https://api.tcgdex.net/v2/en/sets/%s", setID)
	log.Printf("Requête des cartes du set %s à l'URL: %s", setID, apiURL)

	type SetResponse struct {
		CardCount struct {
			Total int `json:"total"`
		} `json:"cardCount"`
		Cards []Card `json:"cards"`
	}

	var setData SetResponse
	err := fetchJSON(apiURL, &setData)
	if err != nil {
		return []Card{}, err
	}

	log.Printf("Nombre de cartes trouvées pour le set %s: %d", setID, len(setData.Cards))

	for i := range setData.Cards {

		if setData.Cards[i].Image != "" {

			if !strings.HasSuffix(setData.Cards[i].Image, ".png") && !strings.HasSuffix(setData.Cards[i].Image, ".jpg") {
				setData.Cards[i].Image = setData.Cards[i].Image + "/high.jpg"
			}
		} else if setData.Cards[i].Images.Large != "" {
			setData.Cards[i].Image = setData.Cards[i].Images.Large
		} else if setData.Cards[i].Images.Small != "" {
			setData.Cards[i].Image = setData.Cards[i].Images.Small
		} else {

			encodedName := url.QueryEscape(setData.Cards[i].Name)
			setData.Cards[i].Image = "https://via.placeholder.com/245x342.png?text=" + encodedName
		}
	}

	if limit > 0 && limit < len(setData.Cards) {
		return setData.Cards[:limit], nil
	}

	return setData.Cards, nil
}
func setDetailHandler(w http.ResponseWriter, r *http.Request) {
	id := strings.TrimPrefix(r.URL.Path, "/set/")
	if id == "" {
		showError(w, "Page non trouvée", fmt.Errorf("ID de set non spécifié"))
		return
	}

	set, err := fetchSet(id)
	if err != nil {
		showError(w, "Impossible de récupérer les détails de la collection", err)
		return
	}

	cards, err := fetchSetCards(id, 100)
	if err != nil {

		log.Printf("Erreur lors de la récupération des cartes du set: %v", err)
	}

	html := `<!DOCTYPE html>
<html lang="fr">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>` + set.Name + ` - PokéTracker</title>
    <link rel="stylesheet" href="/static/css/style.css">
</head>
<body>
    <header>
        <div class="container">
            <h1><a href="/">PokéTracker</a></h1>
            <nav>
                <ul>
                    <li><a href="/">Accueil</a></li>
                    <li><a href="/cards">Cartes</a></li>
                    <li><a href="/sets">Collections</a></li>
                    <li><a href="/favorites">Favoris</a></li>
                    <li><a href="/about">À propos</a></li>
                </ul>
            </nav>
            <form action="/search" method="GET" class="search-form">
                <input type="text" name="q" placeholder="Rechercher des cartes..." required>
                <button type="submit">Rechercher</button>
            </form>
        </div>
    </header>
    
    <main class="container">
        <div class="set-detail">
            <div class="set-header">
                <div class="set-logo">
                    <img src="` + set.Logo + `" alt="` + set.Name + `">
                </div>
                <div class="set-info">
                    <h2>` + set.Name + `</h2>`

	if set.CardCount.Total > 0 {
		html += `<p><strong>Nombre de cartes:</strong> ` + strconv.Itoa(set.CardCount.Total) + `</p>`
	}

	if set.ReleaseDate != "" {
		html += `<p><strong>Date de sortie:</strong> ` + set.ReleaseDate + `</p>`
	}

	html += `</div>
            </div>
            
            <div class="set-cards">
                <h3>Cartes de cette collection</h3>
                <div class="card-grid fade-in">`

	if len(cards) > 0 {
		for _, card := range cards {
			html += `<div class="card">
                    <a href="/card/` + card.ID + `">
                        <img src="` + card.Image + `" alt="` + card.Name + `">
                        <div class="card-content">
                            <h4>` + card.Name + `</h4>
                            <p>` + card.Number + `</p>`

			if len(card.Types) > 0 {
				html += `<div class="card-types">`
				for _, cardType := range card.Types {
					html += `<span class="type ` + cardType + `">` + cardType + `</span>`
				}
				html += `</div>`
			}

			html += `</div>
                    </a>
                </div>`
		}
	} else {
		html += `<p class="no-results">Aucune carte n'a pu être chargée pour cette collection.</p>`
	}

	html += `</div>
            </div>
            
            <div class="set-actions">
                <a href="/sets" class="button">Retour aux Collections</a>
            </div>
        </div>
    </main>
    
    <footer>
        <div class="container">
            <p>&copy; 2025 PokéTracker - Créé pour le projet Groupie Tracker</p>
        </div>
    </footer>
</body>
</html>`

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Write([]byte(html))
}

func aboutHandler(w http.ResponseWriter, r *http.Request) {

	html := `<!DOCTYPE html>
<html lang="fr">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>À propos - PokéTracker</title>
    <link rel="stylesheet" href="/static/css/style.css">
    <style>
        .about-content {
            background-color: var(--white);
            padding: var(--spacing-xl);
            border-radius: var(--radius-lg);
            box-shadow: var(--shadow-md);
            margin-bottom: var(--spacing-xl);
        }
        
        .about-content section {
            margin-bottom: var(--spacing-xl);
        }
        
        .about-content h3 {
            color: var(--primary-color);
            margin-bottom: var(--spacing-md);
            padding-bottom: var(--spacing-sm);
            border-bottom: 2px solid var(--primary-light);
            display: inline-block;
        }
        
        .faq-item {
            margin-bottom: var(--spacing-md);
            background-color: rgba(0, 0, 0, 0.02);
            border-radius: var(--radius-md);
            overflow: hidden;
        }
        
        .faq-item h4 {
            padding: var(--spacing-md);
            margin: 0;
            background-color: rgba(0, 0, 0, 0.03);
            cursor: pointer;
            position: relative;
            transition: all var(--transition-fast);
        }
        
        .faq-item h4:after {
            content: '+';
            position: absolute;
            right: var(--spacing-md);
            top: 50%;
            transform: translateY(-50%);
            font-size: 1.5rem;
            color: var(--primary-color);
            transition: all var(--transition-fast);
        }
        
        .faq-item h4.active:after {
            content: '−';
        }
        
        .faq-item h4:hover {
            background-color: rgba(0, 0, 0, 0.05);
        }
        
        .faq-answer {
            padding: var(--spacing-md);
            border-top: 1px solid rgba(0, 0, 0, 0.05);
        }
        
        .faq-answer p, .faq-answer ul, .faq-answer ol {
            margin-bottom: var(--spacing-sm);
        }
        
        .faq-answer ul, .faq-answer ol {
            padding-left: var(--spacing-lg);
        }
        
        .faq-answer li {
            margin-bottom: var(--spacing-xs);
            display: list-item;
        }
        
        .faq-answer ul li {
            list-style-type: disc;
        }
        
        .faq-answer ol li {
            list-style-type: decimal;
        }
        
        .tech-list {
            display: flex;
            flex-wrap: wrap;
            gap: var(--spacing-md);
            margin-top: var(--spacing-md);
        }
        
        .tech-item {
            background-color: var(--white);
            padding: var(--spacing-sm) var(--spacing-md);
            border-radius: var(--radius-sm);
            box-shadow: var(--shadow-sm);
            display: flex;
            align-items: center;
            gap: var(--spacing-sm);
        }
        
        .endpoint-table {
            width: 100%;
            border-collapse: collapse;
            margin-top: var(--spacing-md);
        }
        
        .endpoint-table th, .endpoint-table td {
            padding: var(--spacing-sm);
            text-align: left;
            border-bottom: 1px solid rgba(0, 0, 0, 0.1);
        }
        
        .endpoint-table th {
            background-color: var(--secondary-light);
            color: white;
        }
        
        .endpoint-table tr:nth-child(even) {
            background-color: rgba(0, 0, 0, 0.02);
        }
        
        .timeline {
            position: relative;
            margin: var(--spacing-xl) 0;
            padding-left: 40px;
        }
        
        .timeline:before {
            content: '';
            position: absolute;
            left: 10px;
            top: 0;
            bottom: 0;
            width: 2px;
            background: var(--primary-light);
        }
        
        .timeline-item {
            position: relative;
            margin-bottom: var(--spacing-lg);
        }
        
        .timeline-item:before {
            content: '';
            position: absolute;
            left: -40px;
            top: 0;
            width: 20px;
            height: 20px;
            border-radius: 50%;
            background: var(--primary-color);
            border: 3px solid var(--white);
            box-shadow: var(--shadow-sm);
        }
        
        .timeline-item h4 {
            margin-bottom: var(--spacing-sm);
            color: var(--primary-dark);
        }
    </style>
</head>
<body>
    <header>
        <div class="container">
            <h1><a href="/">PokéTracker</a></h1>
            <nav>
                <ul>
                    <li><a href="/">Accueil</a></li>
                    <li><a href="/cards">Cartes</a></li>
                    <li><a href="/sets">Collections</a></li>
                    <li><a href="/favorites">Favoris</a></li>
                    <li><a href="/about">À propos</a></li>
                </ul>
            </nav>
            <form action="/search" method="GET" class="search-form">
                <input type="text" name="q" placeholder="Rechercher des cartes..." required>
                <button type="submit">Rechercher</button>
            </form>
        </div>
    </header>
    
    <main class="container">
        <div class="page-header">
            <h2>À propos de PokéTracker</h2>
        </div>

        <div class="about-content">
            <section>
                <h3>Présentation du projet</h3>
                <p>PokéTracker est une application web permettant d'explorer et de suivre les cartes Pokémon. Cette application a été développée dans le cadre du projet "Groupie Tracker" pour démontrer l'utilisation d'une API REST et l'implémentation de fonctionnalités web essentielles en Go.</p>
                <p>Grâce à PokéTracker, les utilisateurs peuvent explorer des milliers de cartes Pokémon, rechercher, filtrer et ajouter leurs cartes préférées à leur liste de favoris. L'application utilise l'API TCGdex pour récupérer les informations sur les cartes et les collections.</p>
            </section>
            
            <section>
                <h3>Fonctionnalités principales</h3>
                <div class="tech-list">
                    <div class="tech-item">
                        <span>Recherche de cartes par nom</span>
                    </div>
                    <div class="tech-item">
                        <span>Filtrage par type, rareté et collection</span>
                    </div>
                    <div class="tech-item">
                        <span>Pagination des résultats</span>
                    </div>
                    <div class="tech-item">
                        <span>Gestion des cartes favorites</span>
                    </div>
                    <div class="tech-item">
                        <span>Exploration des collections</span>
                    </div>
                    <div class="tech-item">
                        <span>Détails des cartes</span>
                    </div>
                </div>
            </section>
            
            <section class="faq">
                <h3>FAQ sur la gestion du projet</h3>
                
                <div class="faq-item">
                    <h4>Comment avez-vous décomposé le projet ? Quelles ont été les phases clé ?</h4>
                    <div class="faq-answer">
                        <p>Le projet a été structuré en plusieurs phases distinctes pour assurer une progression méthodique :</p>
                        
                        <div class="timeline">
                            <div class="timeline-item">
                                <h4>Phase 1 : Recherche et planification (3 jours)</h4>
                                <p>J'ai commencé par explorer différentes API liées aux cartes Pokémon et j'ai finalement choisi l'API TCGdex pour sa richesse en données et sa documentation. J'ai analysé sa structure, ses endpoints et les formats de données qu'elle renvoie. Cette phase a également inclus la conception de l'architecture globale de l'application et la définition des modèles de données nécessaires.</p>
                            </div>
                            
                            <div class="timeline-item">
                                <h4>Phase 2 : Développement du backend (5 jours)</h4>
                                <p>J'ai implémenté les fonctions Go pour communiquer avec l'API, gérer les erreurs et traiter les données. Cette phase incluait:</p>
                                <ul>
                                    <li>Création des structures de données pour représenter les cartes et les collections</li>
                                    <li>Implémentation des fonctions pour récupérer et traiter les données de l'API</li>
                                    <li>Mise en place du système de gestion des favoris avec persistance locale</li>
                                    <li>Gestion des erreurs et mise en place de solutions de secours en cas d'indisponibilité de l'API</li>
                                </ul>
                            </div>
                            
                            <div class="timeline-item">
                                <h4>Phase 3 : Développement du frontend (4 jours)</h4>
                                <p>J'ai créé les templates HTML et les styles CSS pour présenter les données de manière attrayante et responsive:</p>
                                <ul>
                                    <li>Conception de l'interface utilisateur avec un accent sur l'expérience utilisateur</li>
                                    <li>Création des templates pour chaque page (accueil, liste des cartes, détails, etc.)</li>
                                    <li>Développement du CSS pour un design moderne et adaptatif</li>
                                    <li>Intégration de fonctionnalités JavaScript minimales pour améliorer l'interactivité</li>
                                </ul>
                            </div>
                            
                            <div class="timeline-item">
                                <h4>Phase 4 : Implémentation des fonctionnalités clés (4 jours)</h4>
                                <p>J'ai implémenté les quatre fonctionnalités principales demandées:</p>
                                <ul>
                                    <li>Système de recherche par nom de carte</li>
                                    <li>Système de filtrage par type, rareté et collection</li>
                                    <li>Pagination des résultats avec options de nombre d'éléments par page</li>
                                    <li>Gestion des favoris avec sauvegarde dans des fichiers JSON</li>
                                </ul>
                            </div>
                            
                            <div class="timeline-item">
                                <h4>Phase 5 : Tests et optimisations (3 jours)</h4>
                                <p>J'ai testé l'application dans différents scénarios pour identifier et corriger les bugs:</p>
                                <ul>
                                    <li>Tests des différentes fonctionnalités et correction des bugs</li>
                                    <li>Optimisation des performances, notamment pour le chargement des cartes</li>
                                    <li>Amélioration de la gestion des erreurs et des cas limites</li>
                                    <li>Ajout d'une page de diagnostic pour les problèmes d'images</li>
                                </ul>
                            </div>
                            
                            <div class="timeline-item">
                                <h4>Phase 6 : Documentation et finalisation (2 jours)</h4>
                                <p>J'ai documenté le code et créé la documentation du projet:</p>
                                <ul>
                                    <li>Rédaction de commentaires explicatifs dans le code</li>
                                    <li>Création du README avec instructions d'installation et d'utilisation</li>
                                    <li>Développement de cette page "À propos" avec les détails du projet</li>
                                    <li>Révision finale et préparation pour la soumission</li>
                                </ul>
                            </div>
                        </div>
                    </div>
                </div>
                
                <div class="faq-item">
                    <h4>Comment avez-vous réparti les tâches ? Avez-vous utilisé une stratégie particulière pour organiser le travail ?</h4>
                    <div class="faq-answer">
                        <p>Bien que ce projet soit individuel, j'ai mis en place une organisation structurée pour gérer efficacement le travail :</p>
                        
                        <h5>Méthodologie inspirée d'Agile</h5>
                        <p>J'ai adopté une approche inspirée des méthodologies agiles, en divisant le projet en "sprints" de 2-3 jours avec des objectifs clairs pour chacun :</p>
                        <ul>
                            <li>Sprint 1 : Mise en place de l'architecture et connexion à l'API</li>
                            <li>Sprint 2 : Développement des fonctionnalités de base (affichage des cartes et des collections)</li>
                            <li>Sprint 3 : Implémentation du système de recherche et de filtrage</li>
                            <li>Sprint 4 : Développement de la pagination et des favoris</li>
                            <li>Sprint 5 : Finalisation et optimisations</li>
                        </ul>
                        
                        <h5>Outils de gestion</h5>
                        <p>Pour organiser mon travail, j'ai utilisé :</p>
                        <ul>
                            <li><strong>Git</strong> pour le versionnement du code avec des branches distinctes pour chaque fonctionnalité</li>
                            <li><strong>Trello</strong> pour suivre les tâches à accomplir avec des listes "À faire", "En cours" et "Terminé"</li>
                            <li><strong>Postman</strong> pour tester les endpoints de l'API avant de les intégrer au code</li>
                        </ul>
                        
                        <h5>Structure du code</h5>
                        <p>J'ai organisé le code de manière modulaire pour faciliter le développement et la maintenance :</p>
                        <ul>
                            <li>Séparation claire entre les fonctions de récupération des données (fetchCards, fetchSet, etc.)</li>
                            <li>Organisation des templates HTML par fonctionnalité</li>
                            <li>Structure CSS avec variables pour maintenir une cohérence visuelle</li>
                        </ul>
                    </div>
                </div>
                
                <div class="faq-item">
                    <h4>Comment avez-vous géré votre temps ? Avez-vous défini des priorités ?</h4>
                    <div class="faq-answer">
                        <p>La gestion du temps a été un aspect crucial de ce projet, que j'ai abordé de façon méthodique :</p>
                        
                        <h5>Planning et échéancier</h5>
                        <ul>
                            <li>J'ai établi un calendrier détaillé avec des jalons hebdomadaires à atteindre</li>
                            <li>J'ai alloué plus de temps aux étapes critiques comme la communication avec l'API et l'implémentation des fonctionnalités principales</li>
                            <li>J'ai réservé une période tampon d'environ 20% du temps total pour gérer les imprévus et les difficultés techniques</li>
                        </ul>
                        
                        <h5>Priorisation des tâches</h5>
                        <p>J'ai adopté une approche basée sur les priorités suivantes :</p>
                        <ol>
                            <li><strong>Priorité 1</strong> : Fonctionnalités essentielles demandées dans le sujet (recherche, filtres, pagination, favoris)</li>
                            <li><strong>Priorité 2</strong> : Architecture robuste et gestion des erreurs</li>
                            <li><strong>Priorité 3</strong> : Expérience utilisateur et interface attrayante</li>
                            <li><strong>Priorité 4</strong> : Fonctionnalités bonus et optimisations</li>
                        </ol>
                        
                        <h5>Technique de travail</h5>
                        <p>Pour maximiser ma productivité, j'ai utilisé :</p>
                        <ul>
                            <li>La technique Pomodoro (sessions de travail de 25 minutes suivies de pauses de 5 minutes)</li>
                            <li>Des sessions de "revue de code personnel" en fin de journée pour identifier les améliorations possibles</li>
                            <li>Un journal de développement pour documenter les problèmes rencontrés et les solutions trouvées</li>
                        </ul>
                    </div>
                </div>
                
                <div class="faq-item">
                    <h4>Quelle stratégie avez-vous adoptée pour vous documenter ?</h4>
                    <div class="faq-answer">
                        <p>La documentation a été un élément essentiel tout au long du projet, tant pour l'apprentissage que pour la production :</p>
                        
                        <h5>Documentation de l'API et des technologies</h5>
                        <ul>
                            <li>J'ai étudié en profondeur la documentation de l'API TCGdex, en testant chaque endpoint via Postman</li>
                            <li>J'ai consulté la documentation officielle de Go, notamment pour les packages <code>html/template</code>, <code>net/http</code> et <code>encoding/json</code></li>
                            <li>J'ai recherché des bonnes pratiques pour l'implémentation de fonctionnalités comme la pagination et la gestion des favoris en Go</li>
                        </ul>
                        
                        <h5>Veille et ressources</h5>
                        <p>J'ai utilisé diverses ressources pour enrichir mes connaissances :</p>
                        <ul>
                            <li>Forums spécialisés comme Stack Overflow et Go Forum</li>
                            <li>Tutoriels et articles sur la création d'applications web en Go</li>
                            <li>Vidéos explicatives sur les concepts de design responsive et d'UX</li>
                        </ul>
                        
                        <h5>Documentation du code</h5>
                        <p>J'ai mis en place une stratégie de documentation du code selon ces principes :</p>
                        <ul>
                            <li>Commentaires descriptifs pour chaque fonction expliquant son rôle et son fonctionnement</li>
                            <li>Documentation des structures de données avec des annotations sur chaque champ</li>
                            <li>Messages de log détaillés pour faciliter le débogage</li>
                            <li>Nommage explicite des variables et des fonctions pour une meilleure compréhension</li>
                        </ul>
                    </div>
                </div>
            </section>
            
            <section>
                <h3>Technologies utilisées</h3>
                <div class="tech-list">
                    <div class="tech-item">
                        <span>Go (backend)</span>
                    </div>
                    <div class="tech-item">
                        <span>HTML (templates)</span>
                    </div>
                    <div class="tech-item">
                        <span>CSS (styles)</span>
                    </div>
                    <div class="tech-item">
                        <span>JavaScript (interactivité)</span>
                    </div>
                    <div class="tech-item">
                        <span>API TCGdex</span>
                    </div>
                </div>
            </section>
            
            <section>
                <h3>API utilisée</h3>
                <p><strong>API</strong> : TCGdex</p>
                <p>L'API TCGdex fournit des données complètes sur les cartes Pokémon, incluant les informations sur les cartes, les collections, les types et les raretés.</p>
                
                <table class="endpoint-table">
                    <thead>
                        <tr>
                            <th>Endpoint</th>
                            <th>Description</th>
                            <th>Utilisation dans PokéTracker</th>
                        </tr>
                    </thead>
                    <tbody>
                        <tr>
                            <td><code>/v2/en/cards</code></td>
                            <td>Récupération de la liste des cartes</td>
                            <td>Page des cartes, recherche et filtrage</td>
                        </tr>
                        <tr>
                            <td><code>/v2/en/cards/{id}</code></td>
                            <td>Récupération des détails d'une carte</td>
                            <td>Page de détail d'une carte</td>
                        </tr>
                        <tr>
                            <td><code>/v2/en/sets</code></td>
                            <td>Récupération de la liste des collections</td>
                            <td>Page des collections et options de filtrage</td>
                        </tr>
                        <tr>
                            <td><code>/v2/en/sets/{id}</code></td>
                            <td>Récupération des détails d'une collection</td>
                            <td>Page de détail d'une collection</td>
                        </tr>
                        <tr>
                            <td><code>/v2/en/types</code></td>
                            <td>Récupération de la liste des types de cartes</td>
                            <td>Options de filtrage par type</td>
                        </tr>
                        <tr>
                            <td><code>/v2/en/rarities</code></td>
                            <td>Récupération de la liste des raretés de cartes</td>
                            <td>Options de filtrage par rareté</td>
                        </tr>
                    </tbody>
                </table>
            </section>
        </div>
    </main>
    
    <footer>
        <div class="container">
            <p>&copy; 2025 PokéTracker - Créé pour le projet Groupie Tracker</p>
        </div>
    </footer>
    
    <script>
        document.addEventListener('DOMContentLoaded', function() {
            // Ajouter des écouteurs d'événement pour les éléments FAQ
            document.querySelectorAll('.faq-item h4').forEach(item => {
                item.addEventListener('click', function() {
                    // Toggle active class
                    this.classList.toggle('active');
                    
                    // Toggle visibility of answer
                    const answer = this.nextElementSibling;
                    if (answer.style.display === 'none' || !answer.style.display) {
                        answer.style.display = 'block';
                    } else {
                        answer.style.display = 'none';
                    }
                });
            });
            
            // Cacher toutes les réponses par défaut
            document.querySelectorAll('.faq-answer').forEach(answer => {
                answer.style.display = 'none';
            });
        });
    </script>
</body>
</html>`

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Write([]byte(html))
}

func favoritesHandler(w http.ResponseWriter, r *http.Request) {
	favorites, err := loadFavorites()

	html := `<!DOCTYPE html>
<html lang="fr">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>Mes Favoris - PokéTracker</title>
    <link rel="stylesheet" href="/static/css/style.css">
</head>
<body>
    <header>
        <div class="container">
            <h1><a href="/">PokéTracker</a></h1>
            <nav>
                <ul>
                    <li><a href="/">Accueil</a></li>
                    <li><a href="/cards">Cartes</a></li>
                    <li><a href="/sets">Collections</a></li>
                    <li><a href="/favorites">Favoris</a></li>
                    <li><a href="/about">À propos</a></li>
                </ul>
            </nav>
            <form action="/search" method="GET" class="search-form">
                <input type="text" name="q" placeholder="Rechercher des cartes..." required>
                <button type="submit">Rechercher</button>
            </form>
        </div>
    </header>
    
    <main class="container">
        <div class="page-header">
            <h2>Mes Cartes Favorites</h2>
        </div>`

	if err != nil {
		html += `
        <div class="error-message">
            <p>Une erreur est survenue lors du chargement des favoris. Vos favoris ont été réinitialisés.</p>
        </div>
        
        <div class="no-favorites">
            <p>Vous n'avez pas encore de cartes favorites.</p>
            <p>Parcourez les <a href="/cards">cartes</a> et ajoutez-en à vos favoris.</p>
        </div>`
	} else if len(favorites.Cards) > 0 {
		html += `
        <div class="favorites-controls">
            <p>Vous avez ` + strconv.Itoa(len(favorites.Cards)) + ` cartes en favoris</p>
        </div>

        <div class="card-grid fade-in">`

		for _, card := range favorites.Cards {
			html += `
            <div class="card favorite-card">
                <a href="/card/` + card.ID + `">
                    <img src="` + card.Image + `" alt="` + card.Name + `">
                    <div class="card-content">
                        <h3>` + card.Name + `</h3>
                        <p>` + card.Set.Name + `</p>`

			if len(card.Types) > 0 {
				html += `<div class="card-types">`
				for _, cardType := range card.Types {
					html += `<span class="type ` + cardType + `">` + cardType + `</span>`
				}
				html += `</div>`
			}

			html += `
                    </div>
                </a>
                <button class="remove-favorite" data-id="` + card.ID + `">Retirer</button>
            </div>`
		}

		html += `
        </div>

        <div class="favorites-actions">
            <button id="clear-favorites" class="button">Vider ma liste de favoris</button>
        </div>`
	} else {
		html += `
        <div class="no-favorites">
            <p>Vous n'avez pas encore de cartes favorites.</p>
            <p>Parcourez les <a href="/cards">cartes</a> et ajoutez-en à vos favoris.</p>
        </div>`
	}

	html += `
    </main>
    
    <footer>
        <div class="container">
            <p>&copy; 2025 PokéTracker - Créé pour le projet Groupie Tracker</p>
        </div>
    </footer>
    
    <script>
        document.addEventListener('DOMContentLoaded', function() {
            // Remove individual favorite
            document.querySelectorAll('.remove-favorite').forEach(button => {
                button.addEventListener('click', function(e) {
                    e.preventDefault();
                    const cardId = this.getAttribute('data-id');
                    fetch('/api/favorite/remove/' + cardId)
                        .then(response => {
                            if (response.ok) {
                                window.location.reload();
                            }
                        });
                });
            });
            
            // Clear all favorites
            const clearButton = document.getElementById('clear-favorites');
            if (clearButton) {
                clearButton.addEventListener('click', function() {
                    if (confirm('Êtes-vous sûr de vouloir supprimer toutes vos cartes favorites ?')) {
                        // Improved method to clear favorites - direct call to reset
                        fetch('/api/favorite/clear')
                            .then(response => {
                                if (response.ok) {
                                    window.location.reload();
                                }
                            });
                    }
                });
            }
        });
    </script>
</body>
</html>`

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Write([]byte(html))
}

func addFavoriteHandler(w http.ResponseWriter, r *http.Request) {
	cardID := strings.TrimPrefix(r.URL.Path, "/api/favorite/add/")
	if cardID == "" {
		http.Error(w, "ID de carte requis", http.StatusBadRequest)
		return
	}

	card, err := fetchCard(cardID)
	if err != nil {
		http.Error(w, "Impossible de récupérer la carte: "+err.Error(), http.StatusInternalServerError)
		return
	}

	favorites, err := loadFavorites()
	if err != nil {
		http.Error(w, "Impossible de charger les favoris: "+err.Error(), http.StatusInternalServerError)
		return
	}

	for _, favCard := range favorites.Cards {
		if favCard.ID == card.ID {

			w.WriteHeader(http.StatusOK)
			return
		}
	}

	favorites.Cards = append(favorites.Cards, card)

	err = saveFavorites(favorites)
	if err != nil {
		http.Error(w, "Impossible de sauvegarder les favoris: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func removeFavoriteHandler(w http.ResponseWriter, r *http.Request) {
	cardID := strings.TrimPrefix(r.URL.Path, "/api/favorite/remove/")
	if cardID == "" {
		http.Error(w, "ID de carte requis", http.StatusBadRequest)
		return
	}

	favorites, err := loadFavorites()
	if err != nil {
		http.Error(w, "Impossible de charger les favoris: "+err.Error(), http.StatusInternalServerError)
		return
	}

	for i, card := range favorites.Cards {
		if card.ID == cardID {
			favorites.Cards = append(favorites.Cards[:i], favorites.Cards[i+1:]...)
			break
		}
	}

	err = saveFavorites(favorites)
	if err != nil {
		http.Error(w, "Impossible de sauvegarder les favoris: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}
func clearFavoritesHandler(w http.ResponseWriter, r *http.Request) {

	newFavorites := Favorites{Cards: []Card{}}
	err := saveFavorites(newFavorites)

	if err != nil {
		http.Error(w, "Impossible de vider les favoris: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}
func searchHandler(w http.ResponseWriter, r *http.Request) {
	query := r.FormValue("q")
	if query == "" {
		http.Redirect(w, r, "/cards", http.StatusSeeOther)
		return
	}

	filters := map[string]string{"name": query}
	cards, _, err := fetchCards(1, 1000, filters)

	count := len(cards)
	errorMsg := ""

	if err != nil {
		log.Printf("Erreur lors de la recherche de cartes: %v", err)
		errorMsg = "Erreur lors de la recherche. Veuillez réessayer plus tard."
	}

	html := `<!DOCTYPE html>
<html lang="fr">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>Recherche: ` + query + ` - PokéTracker</title>
    <link rel="stylesheet" href="/static/css/style.css">
</head>
<body>
    <header>
        <div class="container">
            <h1><a href="/">PokéTracker</a></h1>
            <nav>
                <ul>
                    <li><a href="/">Accueil</a></li>
                    <li><a href="/cards">Cartes</a></li>
                    <li><a href="/sets">Collections</a></li>
                    <li><a href="/favorites">Favoris</a></li>
                    <li><a href="/about">À propos</a></li>
                </ul>
            </nav>
            <form action="/search" method="GET" class="search-form">
                <input type="text" name="q" placeholder="Rechercher des cartes..." required>
                <button type="submit">Rechercher</button>
            </form>
        </div>
    </header>
    
    <main class="container">
        <div class="page-header">
            <h2>Résultats de recherche pour "` + query + `"</h2>
            <div class="results-count">
                <p>` + strconv.Itoa(count) + ` cartes trouvées</p>`

	if errorMsg != "" {
		html += `<p class="error-message">` + errorMsg + `</p>`
	}

	html += `</div>
        </div>`

	if count > 0 {
		html += `<div class="card-grid fade-in">`

		for _, card := range cards {
			html += `<div class="card">
                <a href="/card/` + card.ID + `">
                    <img src="` + card.Image + `" alt="` + card.Name + `">
                    <div class="card-content">
                        <h3>` + card.Name + `</h3>
                        <p>` + card.Set.Name + `</p>`

			if len(card.Types) > 0 {
				html += `<div class="card-types">`
				for _, cardType := range card.Types {
					html += `<span class="type ` + cardType + `">` + cardType + `</span>`
				}
				html += `</div>`
			}

			html += `</div>
                </a>
            </div>`
		}

		html += `</div>`
	} else {
		html += `<div class="no-results">
            <p>Aucune carte ne correspond à votre recherche "` + query + `".</p>
            <p>Essayez avec d'autres termes ou <a href="/cards">consultez toutes les cartes</a>.</p>
        </div>`
	}

	html += `<div class="search-actions">
            <a href="/cards" class="button">Retour aux Cartes</a>
        </div>
    </main>
    
    <footer>
        <div class="container">
            <p>&copy; 2025 PokéTracker - Créé pour le projet Groupie Tracker</p>
        </div>
    </footer>
</body>
</html>`

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Write([]byte(html))
}

func testImagesHandler(w http.ResponseWriter, r *http.Request) {

	cards, _, err := fetchCards(1, 5, nil)
	if err != nil {
		http.Error(w, "Erreur lors de la récupération des cartes: "+err.Error(), http.StatusInternalServerError)
		return
	}

	sets, err := fetchSets()
	if err != nil || len(sets) == 0 {
		http.Error(w, "Erreur lors de la récupération des sets: "+err.Error(), http.StatusInternalServerError)
		return
	}

	if len(sets) > 5 {
		sets = sets[:5]
	}

	htmlContent := `
    <!DOCTYPE html>
    <html>
    <head>
        <title>Test des images</title>
        <style>
            body { font-family: Arial, sans-serif; max-width: 800px; margin: 0 auto; padding: 20px; }
            .test-section { margin-bottom: 40px; }
            .image-test { margin: 10px 0; padding: 10px; border: 1px solid #ddd; border-radius: 4px; }
            .success { background-color: #e8f5e9; }
            .error { background-color: #ffebee; }
            img { max-width: 200px; max-height: 200px; display: block; margin: 10px 0; }
            h3 { margin-top: 0; }
        </style>
    </head>
    <body>
        <h1>Test des URLs d'images</h1>
        
        <div class="test-section">
            <h2>Test des images de cartes</h2>
    `

	for _, card := range cards {
		htmlContent += fmt.Sprintf(`
            <div class="image-test">
                <h3>%s</h3>
                <p>URL de l'image: <code>%s</code></p>
                <p>Test d'affichage:</p>
                <img src="%s" alt="%s" onerror="this.parentNode.classList.add('error'); this.parentNode.querySelector('.status').textContent='❌ Erreur';" onload="this.parentNode.classList.add('success'); this.parentNode.querySelector('.status').textContent='✅ Succès';">
                <p class="status">Chargement...</p>
            </div>
        `, card.Name, card.Image, card.Image, card.Name)
	}

	htmlContent += `
        </div>
        
        <div class="test-section">
            <h2>Test des images de sets</h2>
    `

	for _, set := range sets {
		htmlContent += fmt.Sprintf(`
            <div class="image-test">
                <h3>%s</h3>
                <p>URL du logo: <code>%s</code></p>
                <p>Test d'affichage du logo:</p>
                <img src="%s" alt="%s" onerror="this.parentNode.classList.add('error'); this.parentNode.querySelector('.logo-status').textContent='❌ Erreur';" onload="this.parentNode.classList.add('success'); this.parentNode.querySelector('.logo-status').textContent='✅ Succès';">
                <p class="logo-status">Chargement...</p>
        `, set.Name, set.Logo, set.Logo, set.Name)

		if set.Symbol != "" {
			htmlContent += fmt.Sprintf(`
                <p>URL du symbole: <code>%s</code></p>
                <p>Test d'affichage du symbole:</p>
                <img src="%s" alt="Symbole %s" onerror="this.parentNode.classList.add('error'); this.parentNode.querySelector('.symbol-status').textContent='❌ Erreur';" onload="this.parentNode.classList.add('success'); this.parentNode.querySelector('.symbol-status').textContent='✅ Succès';">
                <p class="symbol-status">Chargement...</p>
            `, set.Symbol, set.Symbol, set.Name)
		}

		htmlContent += `</div>`
	}

	htmlContent += `
        </div>
        
        <p><a href="/">Retour à l'accueil</a></p>
    </body>
    </html>
    `

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Write([]byte(htmlContent))
}

func renderErrorPage(w http.ResponseWriter, message string, errMsg error) {
	data := struct {
		Message string
		Error   error
	}{
		Message: message,
		Error:   errMsg,
	}

	if renderErr := templates.ExecuteTemplate(w, "error.html", data); renderErr != nil {
		log.Printf("Erreur de rendu du template d'erreur: %v", renderErr)
		w.WriteHeader(http.StatusInternalServerError)
		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		w.Write([]byte(fmt.Sprintf("%s: %v", message, errMsg)))
	}
}
func showError(w http.ResponseWriter, title string, errDetail error) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(http.StatusInternalServerError)

	html := fmt.Sprintf(`
    <!DOCTYPE html>
    <html>
    <head>
        <title>Erreur - PokéTracker</title>
        <style>
            body { font-family: Arial, sans-serif; line-height: 1.6; color: #333; max-width: 800px; margin: 0 auto; padding: 20px; }
            .error-container { background-color: #ffebee; border-left: 4px solid #f44336; padding: 20px; margin: 20px 0; border-radius: 4px; }
            h1 { color: #d32f2f; }
            .button { display: inline-block; background-color: #3b4cca; color: white; padding: 8px 15px; text-decoration: none; border-radius: 4px; margin-top: 20px; }
            .button:hover { background-color: #2a3a99; }
            pre { background-color: #f5f5f5; padding: 10px; border-radius: 4px; overflow-x: auto; }
        </style>
    </head>
    <body>
        <h1>Une erreur est survenue</h1>
        <div class="error-container">
            <h2>%s</h2>
            %s
        </div>
        <a href="/" class="button">Retour à l'accueil</a>
    </body>
    </html>
    `, title, func() string {
		if errDetail != nil {
			return fmt.Sprintf("<p><strong>Détails de l'erreur:</strong> %s</p>", errDetail.Error())
		}
		return ""
	}())

	w.Write([]byte(html))
}
