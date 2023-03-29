package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"math/rand"
	"net/http"
	"strconv"
	"time"
)

type StorageProvider struct {
	Providers []Provider `json:"storageProviders"`
}

type ProviderSelect struct {
	Provider Provider `json:"0"`
}
type Provider struct {
	ID                             string `json:"id"`
	Address                        string `json:"address"`
	AddressOfOwner                 string `json:"address_of_owner"`
	AddressOfWorker                string `json:"address_of_worker"`
	AddressOfBeneficiary           string `json:"address_of_beneficiary"`
	SectorSizeBytes                string `json:"sector_size_bytes"`
	MaxPieceSizeBytes              string `json:"max_piece_size_bytes"`
	MinPieceSizeBytes              string `json:"min_piece_size_bytes"`
	PriceAttofil                   string `json:"price_attofil"`
	PriceVerifiedAttofil           string `json:"price_verified_attofil"`
	BalanceAttofil                 string `json:"balance_attofil"`
	LockedFundsAttofil             string `json:"locked_funds_attofil"`
	InitialPledgeAttofil           string `json:"initial_pledge_attofil"`
	RawPowerBytes                  string `json:"raw_power_bytes"`
	QualityAdjustedPowerBytes      string `json:"quality_adjusted_power_bytes"`
	TotalRawPowerBytes             string `json:"total_raw_power_bytes"`
	TotalQualityAdjustedPowerBytes string `json:"total_quality_adjusted_power_bytes"`
	TotalStorageDealCount          string `json:"total_storage_deal_count"`
	TotalSectorsSealedByPostCount  string `json:"total_sectors_sealed_by_post_count"`
	PeerID                         string `json:"peer_id"`
	Height                         string `json:"height"`
	LotusVersion                   string `json:"lotus_version"`
	Multiaddrs                     struct {
		Addresses []string `json:"addresses"`
	} `json:"multiaddrs"`
	Metadata             interface{} `json:"metadata"`
	AddressOfControllers struct {
		Addresses []string `json:"addresses"`
	} `json:"address_of_controllers"`
	Tipset struct {
		Cids []struct {
			NAMING_FAILED string `json:"/"`
		} `json:"cids"`
	} `json:"tipset"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type Pricing struct {
	StoragePrice string `json:"storagePrice"`
}

func fetchProviders(sizeInBytes int64) ([]Provider, error) {
	response, err := http.Get("https://data.storage.market/api/providers")
	if err != nil {
		return nil, err
	}
	defer response.Body.Close()

	body, err := io.ReadAll(response.Body)
	if err != nil {
		return nil, err
	}

	var providers StorageProvider
	err = json.Unmarshal(body, &providers)
	if err != nil {
		return nil, err
	}

	// Filter providers by piece size
	var filteredProviders []Provider
	for _, provider := range providers.Providers {
		pvMinPieceSize, _ := strconv.ParseInt(provider.MinPieceSizeBytes, 10, 64)
		pvMaxPieceSize, _ := strconv.ParseInt(provider.MaxPieceSizeBytes, 10, 64)
		fmt.Println("pvMinPieceSize, pvMaxPieceSize", pvMinPieceSize, pvMaxPieceSize)
		fmt.Println("sizeInBytes", sizeInBytes)
		if pvMinPieceSize <= sizeInBytes && pvMaxPieceSize >= sizeInBytes {
			filteredProviders = append(filteredProviders, provider)
		}
	}

	return filteredProviders, nil
}
func providerInfoHandler(w http.ResponseWriter, r *http.Request) {
	addr := r.URL.Query().Get("addr")

	response, err := http.Get("https://data.storage.market/api/providers/" + addr)
	if err != nil {
		http.Error(w, "Error fetching providers: "+err.Error(), http.StatusInternalServerError)
	}
	defer response.Body.Close()

	body, err := io.ReadAll(response.Body)
	if err != nil {
		http.Error(w, "Error fetching providers: "+err.Error(), http.StatusInternalServerError)
	}

	var providers ProviderSelect
	err = json.Unmarshal(body, &providers)

	if err != nil {
		http.Error(w, "Error fetching providers: "+err.Error(), http.StatusInternalServerError)
	}

	jsonData, err := json.Marshal(providers)
	if err != nil {
		http.Error(w, "Error encoding provider: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write(jsonData)
}
func providersHandler(w http.ResponseWriter, r *http.Request) {
	// Parse query parameters
	sizeInBytes, _ := strconv.ParseInt(r.URL.Query().Get("size_bytes"), 10, 64)
	//minPieceSize, _ := strconv.ParseInt(r.URL.Query().Get("min_piece_size_bytes"), 10, 64)
	//maxPieceSize, _ := strconv.ParseInt(r.URL.Query().Get("max_piece_size_bytes"), 10, 64)

	providers, err := fetchProviders(sizeInBytes)
	if err != nil {
		http.Error(w, "Error fetching providers: "+err.Error(), http.StatusInternalServerError)
		return
	}

	if len(providers) == 0 {
		http.Error(w, "No providers found", http.StatusNotFound)
		return
	}

	// Select one random provider
	rand.Seed(time.Now().UnixNano())
	randomIndex := rand.Intn(len(providers))
	randomProvider := providers[randomIndex]

	jsonData, err := json.Marshal(randomProvider)
	if err != nil {
		http.Error(w, "Error encoding provider: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write(jsonData)
}

func main() {
	http.HandleFunc("/api/providers", providersHandler)
	http.HandleFunc("/api/provider/info", providerInfoHandler)

	fmt.Println("Server started on [::]:8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
