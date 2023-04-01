// It takes an IP address and a list of IP addresses, and returns the IP address from the list that is closest to the given
// IP address
package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"math"
	"math/rand"
	"net/http"
	"regexp"
	"strconv"
	"strings"
	"time"
)

type StorageProvider struct {
	Providers []Provider `json:"storageProviders"`
}

type IpInfoApiResponse struct {
	IP       string `json:"ip"`
	Hostname string `json:"hostname"`
	City     string `json:"city"`
	Region   string `json:"region"`
	Country  string `json:"country"`
	Loc      string `json:"loc"`
	Postal   string `json:"postal"`
	Timezone string `json:"timezone"`
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

func fetchProviders(sizeInBytes int64, sourceIp string) ([]Provider, error) {
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

func ExtractIPAddress(str string) (string, error) {
	re := regexp.MustCompile(`/ip4/(\d{1,3}\.\d{1,3}\.\d{1,3}\.\d{1,3})/tcp/\d+`)
	matches := re.FindStringSubmatch(str)
	if matches == nil {
		return "", fmt.Errorf("could not extract IP address from string")
	}
	return matches[1], nil
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
	sourceIp := r.URL.Query().Get("source_ip")

	providers, err := fetchProviders(sizeInBytes, sourceIp)
	if err != nil {
		http.Error(w, "Error fetching providers: "+err.Error(), http.StatusInternalServerError)
		return
	}

	if len(providers) == 0 {
		http.Error(w, "No providers found", http.StatusNotFound)
		return
	}
	if sourceIp != "" {
		// Filter providers by source IP
		var ips []string
		var providerIps []Provider
		for _, provider := range providers {
			for _, a := range provider.Multiaddrs.Addresses {
				ip, err := ExtractIPAddress(a)
				if err != nil {
					// Handle error
				}

				ips = append(ips, ip)
				providerIps = append(providerIps, provider)
			}
		}
		nearestFromTheIp, err := NearestIPCoordinate(sourceIp, ips)
		if err != nil {
			http.Error(w, "Error fetching nearest IP Coordinates: "+err.Error(), http.StatusInternalServerError)
			return
		}
		jsonData, err := json.Marshal(nearestFromTheIp)
		if err != nil {
			http.Error(w, "Error encoding provider: "+err.Error(), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.Write(jsonData)
	} else {

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
}

func isIPv4(address string) bool {
	return len(address) >= 7 && address[0:7] == "/ip4/"
}
func Haversine(lat1, lon1, lat2, lon2 float64) float64 {
	R := 6371.0 // Earth radius in kilometers

	lat1_rad := math.Pi * lat1 / 180.0
	lon1_rad := math.Pi * lon1 / 180.0
	lat2_rad := math.Pi * lat2 / 180.0
	lon2_rad := math.Pi * lon2 / 180.0

	dlon := lon2_rad - lon1_rad
	dlat := lat2_rad - lat1_rad

	a := math.Sin(dlat/2)*math.Sin(dlat/2) + math.Cos(lat1_rad)*math.Cos(lat2_rad)*math.Sin(dlon/2)*math.Sin(dlon/2)
	c := 2 * math.Atan2(math.Sqrt(a), math.Sqrt(1-a))

	distance := R * c
	return distance
}
func GeolocateIP(ip string) (float64, float64, error) {
	url := fmt.Sprintf("https://ipinfo.io/%s/json?token=<>", ip)
	response, err := http.Get(url)
	if err != nil {
		return 0.0, 0.0, fmt.Errorf("error geolocating IP address %s: %s", ip, err)
	}
	defer response.Body.Close()

	if response.StatusCode == 200 {
		var data IpInfoApiResponse
		err := json.NewDecoder(response.Body).Decode(&data)
		if err != nil {
			return 0.0, 0.0, fmt.Errorf("error parsing response JSON: %s", err)
		}

		latlon := strings.Split(data.Loc, ",")
		latitude, err := strconv.ParseFloat(latlon[0], 64)
		if err != nil {
			return 0.0, 0.0, fmt.Errorf("error parsing latitude: %s", err)
		}

		longitude, err := strconv.ParseFloat(latlon[1], 64)
		if err != nil {
			return 0.0, 0.0, fmt.Errorf("error parsing longitude: %s", err)
		}

		return latitude, longitude, nil
	} else {
		return 0.0, 0.0, fmt.Errorf("error geolocating IP address %s: %s", ip, response.Status)
	}
}

func NearestIPCoordinate(ip string, ips []string) (string, error) {
	givenLatitude, givenLongitude, err := GeolocateIP(ip)
	if err != nil {
		return "", err
	}

	var minDistance float64
	var nearestIP string
	first := true
	for _, otherIP := range ips {
		if otherIP == ip {
			continue
		}

		lat, lon, err := GeolocateIP(otherIP)
		if err != nil {
			continue
		}

		dist := Haversine(givenLatitude, givenLongitude, lat, lon)
		if first || dist < minDistance {
			minDistance = dist
			nearestIP = otherIP
			first = false
		}
	}

	if nearestIP == "" {
		return "", fmt.Errorf("no nearby IP address found for %s", ip)
	}

	return nearestIP, nil
}

func main() {
	http.HandleFunc("/api/providers", providersHandler)
	http.HandleFunc("/api/provider/info", providerInfoHandler)

	fmt.Println("Server started on [::]:8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
