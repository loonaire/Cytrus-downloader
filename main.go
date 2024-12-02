package main

import (
	"cytrusdownloader/cytrus5"
	"cytrusdownloader/cytrus6"
	"flag"
	"fmt"
	"runtime"
	"strings"
)

/*
	type Manifest struct {
		fragments []Fragment
	}

	type Fragment struct {
		name    string
		files   []File
		bundles []Bundle
	}

	type Bundle struct {
		hash   string
		chunks []Chunk
	}

	type Chunk struct {
		hash   string
		size   int64
		offset int64
	}

	type File struct {
		name       string
		size       int64
		hash       string
		chunks     []Chunk
		executable bool
		symlink    string
	}
*/

func main() {
	var game string
	var version string
	var platform string
	var release string
	var manifestFile string
	var outDownload string

	flag.StringVar(&game, "game", "", "Nom du jeu à téléchager (liste non complète) [dofus|retro|wakfu]")
	flag.StringVar(&version, "version", "latest", "Version précise à téléchargée, par défaut la dernière version est téléchargée")
	flag.StringVar(&platform, "platform", runtime.GOOS, "Plateforme choisie, par défaut il s'agit de celle du système d'exploitation [windows,linux,darwin]")
	flag.StringVar(&release, "release", "main", "Version à télécharger, main si n'est pas précisé [main|beta]")
	flag.StringVar(&manifestFile, "manifest-file", "", "Utilise un fichier manifest en local plutot qu'aller le télécharger sur le cdn (Cytrus 6 seulement)")
	flag.StringVar(&outDownload, "outdir", "out/", "Emplacement de sortie du téléchargement")
	flag.Parse()

	// pour éviter les problèmes, on met tout en minuscule
	game = strings.ToLower(game)
	platform = strings.ToLower(platform)
	release = strings.ToLower(release)

	if game == "" {
		gamelist, err := getAvalaibleGameList()
		if err != nil {
			fmt.Println("Erreur, veuillez indiquer le nom d'un jeu")
			return
		}
		fmt.Println("Erreur, veuillez indiquer un jeu, liste des jeux disponibles: ", gamelist)
		return
	} else if version == "latest" {
		gameExist, err := isGameAvalaible(game)
		if err != nil {
			fmt.Println("Erreur, lors de la vérification du jeu")
			return
		}
		if gameExist == false {
			fmt.Println("Le nom du jeu saisi n'existe pas")
			return
		}
	}

	if version == "latest" {
		// on récupère la dernière version
		var err error
		version, err = getLastVersionOfGame(game, platform, release)
		if err != nil {
			fmt.Println("Impossible de vérifier la dernière version disponible du jeu")
			return
		}
	}

	fmt.Println("Informations sur les données à télécharger")
	fmt.Println("Nom du jeu:", game, " plateforme:", platform, " release:", release, " version:", version)

	if strings.HasPrefix(version, "6.0_") {
		if err := cytrus6.Cytrus6Downloader(manifestFile, game, release, platform, version, outDownload); err != nil {
			fmt.Println(err)
			return
		}
		fmt.Println("Le téléchargement s'est correctement terminé")
	} else if strings.HasPrefix(version, "5.0_") {
		fmt.Println("Téléchargement depuis cytrus 5")
		if err := cytrus5.Cytrus5Downloader(manifestFile, game, release, platform, version, outDownload); err != nil {
			fmt.Println(err)
			return
		}
		fmt.Println("Le téléchargement s'est correctement terminé")
	} else {
		fmt.Println("La version de cytrus indiquée est invalide")
	}

}
