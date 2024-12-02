package cytrus5

import (
	"archive/tar"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
)

type File struct { // contenu de "file", le nom du fichier est donné par une map de string
	Hash       string `json:"hash"`
	Size       int64  `json:"size"`
	Executable bool   `json:"executable,omitempty"`
}

type Hash struct {
	Hash []string `json:"hashes"`
	Size int64    `json:"size"`
}

type Fragment struct { // = contenu de "windows" "main", "configuration" "linux" "darwin"
	Files map[string]File `json:"files"`
	Packs map[string]Hash `json:"packs,omitempty"`
}

// les dernières versions dispo sur cette version de cytrus sont accessibles via l'url https://launcher.cdn.ankama.com/cytrus.json

func Cytrus5Downloader(manifestFile string, game string, release string, platform string, version string, outputDir string) error {
	jsonData := []byte{}
	if len(manifestFile) > 5 && strings.HasSuffix(manifestFile, ".json") {
		// si le fichier se termine par .manifest on essaye de l'ouvrir
		file, err := os.ReadFile(manifestFile)
		if err != nil {
			return errors.New("Erreur lors de l'ouverture du fichier manifest")
		}
		jsonData = file

	} else if len(manifestFile) > 1 && !strings.HasSuffix(manifestFile, ".json") {
		return errors.New("Le fichier Manifest n'a pas la bonne extension")
	} else {
		// si le fichier n'est pas indiqué, on télécharge le fichier
		data, errDownloadJson := downloadJsonManifest(game, release, platform, version)
		if errDownloadJson != nil {
			return errors.New("Erreur lors du téléchargement du fichier manifest json " + errDownloadJson.Error())
		}
		jsonData = data
	}

	// charge les informations du json
	jsonUnmarshal := map[string]Fragment{}
	json.Unmarshal(jsonData, &jsonUnmarshal)

	contentDestination := fmt.Sprintf("%s%s/%s/%s", outputDir, game, strings.TrimPrefix(version, "5.0_"), platform)
	var wg sync.WaitGroup

	for k, fragment := range jsonUnmarshal {
		downloadDestination := fmt.Sprintf("%s/%s/", contentDestination, k)
		if errCreateDir := os.MkdirAll(downloadDestination, os.ModePerm); errCreateDir != nil {
			fmt.Println("Impossible de crée le dossier de destination, emplacement:" + downloadDestination + "\n[ERREUR]:" + errCreateDir.Error())
			return errCreateDir
		}
		wg.Add(1)
		go func() {
			defer wg.Done()

			if errCreateDir := os.MkdirAll(downloadDestination, os.ModePerm); errCreateDir != nil {
				fmt.Println("Impossible de crée le dossier de destination, emplacement:" + downloadDestination + "\n[ERREUR]:" + errCreateDir.Error())
				return
			}

			if len(fragment.Packs) > 0 {
				// le fragment contient des Packs, on les télécharges et on les extrait
				for packName := range fragment.Packs {
					// on télécharge le pack
					packFilePath := fmt.Sprintf("%s%s", downloadDestination, packName)
					downloadUrl := fmt.Sprintf("https://launcher.cdn.ankama.com/%s/hashes/%s/%s", game, packName[0:2], packName)
					if err := downloadFile(downloadUrl, packFilePath); err != nil {
						fmt.Println("Erreur lors du téléchargement du pack" + packName + "\n[ERREUR]:" + err.Error())
						break
					}
					fmt.Println("Téléchargement du fichier Pack", packName, "Url:", downloadUrl)
					if err := unpackPackFile(fragment.Files, downloadDestination, packName); err != nil {
						fmt.Println(err)
					}
					// on supprime le fichier, il ne sera plus utiliser
					os.Remove(packFilePath)
				}
			}

			// si le fragment ne contient pas de Packs, on téléchargement les fichiers de manière directe
			for fileName, file := range fragment.Files {
				isFileInPackFile := false
				for _, pack := range fragment.Packs {
					for _, hash := range pack.Hash {
						if hash == file.Hash {
							isFileInPackFile = true
						}
					}
				}

				if isFileInPackFile == false {
					if errCreateDir := os.MkdirAll(fmt.Sprintf("%s%s", downloadDestination, filepath.Dir(fileName)), os.ModePerm); errCreateDir != nil {
						fmt.Println("Impossible de crée le dossier de destination, emplacement:" + downloadDestination + "\n[ERREUR]:" + errCreateDir.Error())
						break
					}
					wg.Add(1)
					go func() {
						defer wg.Done()
						downloadUrl := fmt.Sprintf("https://launcher.cdn.ankama.com/%s/hashes/%s/%s", game, file.Hash[0:2], file.Hash)
						fmt.Println("Téléchargement du fichier", fileName, "URL:", downloadUrl)
						if err := downloadFile(downloadUrl, fmt.Sprintf("%s%s", downloadDestination, fileName)); err != nil {
							fmt.Println("Erreur lors du téléchargement du fichier" + fileName + "\n[ERREUR]: " + err.Error())
						}
					}()

				}
			}

		}()

	}
	wg.Wait()

	return nil
}

func downloadJsonManifest(game string, release string, platform string, version string) ([]byte, error) {
	// télécharge le fichier de manifest
	res, err := http.Get(fmt.Sprintf("https://launcher.cdn.ankama.com/%s/releases/%s/%s/%s.json", game, release, platform, version))
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

func downloadFile(downloadUrl string, destinationFile string) error {
	res, err := http.Get(downloadUrl)
	if err != nil {
		return errors.New("Erreur de lien de telechargement d'un fichier, url: " + downloadUrl + "\n[ERREUR]: " + err.Error())
	}
	defer res.Body.Close()

	file, errOpenFile := os.OpenFile(destinationFile, os.O_RDWR|os.O_CREATE, os.ModePerm)
	if errOpenFile != nil {
		return errors.New("Erreur lors de l'ouverture du fichier " + destinationFile + "\n[ERREUR]: " + errOpenFile.Error())
	}
	defer file.Close()

	_, errCopyToFile := io.Copy(file, res.Body)
	if errCopyToFile != nil {
		return errors.New("Erreur lors de la copie du contenu vers le fichier" + errCopyToFile.Error())
	}
	return nil
}

func unpackPackFile(files map[string]File, fragmentDir string, packName string) error {
	// ouvre le fichier pack en lecteur
	packFilePath := fmt.Sprintf("%s%s", fragmentDir, packName)
	packFileContent, errOpenFile := os.Open(packFilePath)
	if errOpenFile != nil {
		return errors.New("Erreur lors de l'ouverture du fichier " + packFilePath + "\n[ERREUR]: " + errOpenFile.Error())
	}
	defer packFileContent.Close()

	// le contenu du fichier pack est archivé dans du contenu tar
	tarReader := tar.NewReader(packFileContent)

	for {
		header, err := tarReader.Next()
		if err == io.EOF {
			break // il n'y a plus rien à lire
		} else if err != nil {
			return errors.New("Erreur inconnue lors de l'extraction de l'archive")
		}

		switch header.Typeflag {
		case tar.TypeReg:
			// on charge le fichier lié au hash puis on envoi le contenu de l'archive dans son fichier
			for fileName, file := range files {
				if header.Name == file.Hash {
					// le fichier correspond au hash du fichier pack
					destinationFileName := fmt.Sprintf("%s%s", fragmentDir, fileName)
					os.MkdirAll(filepath.Dir(destinationFileName), os.ModePerm) // crée l'arborescence

					outFile, errCreateFile := os.OpenFile(destinationFileName, os.O_RDWR|os.O_CREATE, os.ModePerm)
					if errCreateFile != nil {
						return errors.New("Erreur lors de la création du fichier " + fragmentDir + fileName)
					}
					defer outFile.Close()

					if _, err := io.Copy(outFile, tarReader); err != nil {
						return errors.New("Erreur lors de l'écriture des données dans le fichier")
					}

					fmt.Println("Extraction du fichier", fileName, " depuis le pack", packName)
				}
			}
		default:
		}

	}
	return nil
}
