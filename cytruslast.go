package main

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"strconv"
)

const (
	CYTRUS_LAST_GAMES_VERSION = "https://cytrus.cdn.ankama.com/cytrus.json"
)

type Cytrus struct {
	Version int64           `json:"version"`
	Name    string          `json:"name"`
	Games   map[string]Game `json:"games"`
}

type Game struct {
	Name      string   `json:"name"`
	Order     int64    `json:"order"`
	GameId    int64    `json:"gameId"`
	Assets    Asset    `json:"assets"`
	Platforms Platform `json:"platforms"`
}

type Asset struct {
	Metas Meta `json:"meta"`
}

type Meta struct {
	Beta string `json:"beta"`
	Main string `json:"main"`
}

type Platform struct {
	Linux   VersionsAvalaible `json:"linux,omitempty"`
	Windows VersionsAvalaible `json:"windows,omitempty"`
	Darwin  VersionsAvalaible `json:"darwin,omitempty"`
}

type VersionsAvalaible struct {
	Beta string `json:"beta"`
	Main string `json:"main"`
}

func downloadLastCytrusJson() ([]byte, error) {
	// télécharge le fichier de manifest
	res, err := http.Get(CYTRUS_LAST_GAMES_VERSION)
	if err != nil {
		return []byte{}, errors.New("Erreur lors de la requete du téléchargement du fichier manifest, erreur: " + strconv.FormatInt(int64(res.StatusCode), 10))
	}

	defer res.Body.Close()
	data, errReadBody := io.ReadAll(res.Body)
	if errReadBody != nil {
		return []byte{}, errors.New("Erreur lors de la lecture du corps de la requête")
	}
	return data, nil
}

func getAvalaibleGameList() ([]string, error) {
	lastcytrusjson, err := downloadLastCytrusJson()
	if err != nil {
		return []string{}, err
	}

	lastcytrusUnmarshal := Cytrus{}
	if errorUnmarshal := json.Unmarshal(lastcytrusjson, &lastcytrusUnmarshal); errorUnmarshal != nil {
		return []string{}, errorUnmarshal
	}

	gameList := []string{}
	for gameName := range lastcytrusUnmarshal.Games {
		gameList = append(gameList, gameName)
	}

	return gameList, nil
}

func isGameAvalaible(gameName string) (bool, error) {
	gameExist := false
	lastcytrusjson, err := downloadLastCytrusJson()
	if err != nil {
		return false, err
	}

	lastcytrusUnmarshal := Cytrus{}
	if errorUnmarshal := json.Unmarshal(lastcytrusjson, &lastcytrusUnmarshal); errorUnmarshal != nil {
		return false, errorUnmarshal
	}

	for cytrusGameName := range lastcytrusUnmarshal.Games {
		if gameName == cytrusGameName {
			gameExist = true
			break
		}
	}

	return gameExist, nil
}

func getLastVersionOfGame(gameName string, platform string, release string) (string, error) {

	lastcytrusjson, err := downloadLastCytrusJson()
	if err != nil {
		return "", err
	}

	lastcytrusUnmarshal := Cytrus{}
	if errorUnmarshal := json.Unmarshal(lastcytrusjson, &lastcytrusUnmarshal); errorUnmarshal != nil {
		return "", errorUnmarshal
	}

	releasesAvalaible := VersionsAvalaible{}
	switch platform {
	case "windows":
		releasesAvalaible = lastcytrusUnmarshal.Games[gameName].Platforms.Windows
	case "linux":
		releasesAvalaible = lastcytrusUnmarshal.Games[gameName].Platforms.Linux
	case "darwin":
		releasesAvalaible = lastcytrusUnmarshal.Games[gameName].Platforms.Darwin
	default:
		return "", errors.New("Erreur, la plateforme n'existe pas")
	}

	version := ""
	switch release {
	case "beta":
		version = releasesAvalaible.Beta
	case "main":
		version = releasesAvalaible.Main
	default:
		return "", errors.New("Errur, la release n'existe pas")
	}

	return version, nil

}
