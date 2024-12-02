package cytrus6

import (
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"sync"
)

func Cytrus6Downloader(manifestFile string, game string, release string, platform string, version string, outputDir string) error {
	var manifestData []byte

	if len(manifestFile) > 9 && strings.HasSuffix(manifestFile, ".manifest") {
		// si le fichier se termine par .manifest on essaye de l'ouvrir
		file, err := os.ReadFile(manifestFile)
		if err != nil {
			return errors.New("Erreur lors de l'ouverture du fichier manifest")
		}
		manifestData = file

	} else if len(manifestFile) > 1 && !strings.HasSuffix(manifestFile, ".manifest") {
		return errors.New("Le fichier Manifest n'a pas la bonne extension")
	} else {
		// Telecharge le fichier de manifest de la version souhaitée
		// si le fichier n'est pas saisi ou n'est pas valide, on essaye de télécharger le fichier de manifest
		data, errDownloadManifest := downloadManifest(game, release, platform, version)
		if errDownloadManifest != nil {
			return errors.New("Erreur lors du téléchargement du fichier de manifest " + errDownloadManifest.Error())
		}
		manifestData = data
	}

	// convertis le contenu du fichier manifest en structure de données
	manifestExtracted := extractManifestFromFileData(manifestData)

	// telecharge les fichiers bundles
	contentDestination := fmt.Sprintf("%s%s/%s/%s", outputDir, game, platform, strings.TrimPrefix(version, "6.0_"))

	var wg sync.WaitGroup
	for _, fragment := range manifestExtracted.fragments {
		wg.Add(1)
		go func() {
			defer wg.Done()
			downloadDestination := fmt.Sprintf("%s/%s/", contentDestination, fragment.name)
			os.MkdirAll(downloadDestination, os.ModePerm)
			// parcours les bundle pour l'extraction
			for _, bundle := range fragment.bundles {
				downloadURL := fmt.Sprintf("https://cytrus.cdn.ankama.com/%s/bundles/%s/%s", game, bundle.hash[0:2], bundle.hash)
				bundleFilePath := fmt.Sprintf("%s/%s", downloadDestination, bundle.hash)

				fmt.Println("Telechargement du bundle", bundle.hash, "URL:", downloadURL)
				errDownload := downloadBundleFile(downloadURL, bundleFilePath)
				if errDownload != nil {
					fmt.Println("Erreur lors du téléchargement du fichier bundle " + bundle.hash)
					break
				}
				extractBundleFile(bundle, downloadDestination, fragment.files)
				os.Remove(bundleFilePath)
				fmt.Println("Tous les fichiers ont été téléchargés et extrait dans le répertoire ", downloadDestination)
			}
		}()
	}
	wg.Wait()
	return nil
}

func extractBundleFile(bundle Bundle, downloadDestination string, fragmentFiles []File) error {
	bundleFileContent, errOpenFile := os.Open(fmt.Sprintf("%s/%s", downloadDestination, bundle.hash))
	if errOpenFile != nil {
		return errors.New("Impossible d'ouvrir le fichier bundle" + errOpenFile.Error())
	}
	defer bundleFileContent.Close()

	for _, chunkBundle := range bundle.chunks {
		for _, file := range fragmentFiles {
			if len(file.chunks) == 0 && (chunkBundle.hash == file.hash) {
				// si le fichier n'a pas de chunk, le fichier complet tiens sur un chunk du bundle
				//fmt.Println("Extraction du fichier ", file.name)
				extractChunkToFile(bundleFileContent, chunkBundle, fmt.Sprintf("%s/%s", downloadDestination, file.name), 0)
			}

			if len(file.chunks) > 0 {
				for _, chunkFile := range file.chunks {
					if chunkFile.hash == chunkBundle.hash {
						// on crée le fichier et on copie de contenu du chuck de bundleFileContent dans ce fichier
						//fmt.Println("Extraction d'un chunk du fichier ", file.name)
						extractChunkToFile(bundleFileContent, chunkBundle, fmt.Sprintf("%s/%s", downloadDestination, file.name), chunkFile.offset)
					}
				}
			}
		}
	}
	return nil
}

func extractChunkToFile(bundleFileContent *os.File, bundleChunk Chunk, destinationFile string, fileOffset int64) error {
	// lis le contenu du chunk
	bufferContent := make([]byte, bundleChunk.size)
	_, errReadChunk := bundleFileContent.ReadAt(bufferContent, bundleChunk.offset)
	if errReadChunk != nil {
		return errors.New("Erreur lors de la lecture du chunk")
	}
	// crée le path du fichier
	errMkDir := os.MkdirAll(filepath.Dir(destinationFile), os.ModePerm)
	if errMkDir != nil {
		return errors.New("Erreur lors de la création du répertoire")
	}
	// ouvre le fichier
	finalFile, err := os.OpenFile(destinationFile, os.O_RDWR|os.O_CREATE, os.ModePerm)
	if err != nil {
		return errors.New("Erreur lors de la création ou création du fichier")
	}
	defer finalFile.Close()
	// ecris les ddonnées dans le fichier
	_, errWriteChunk := finalFile.WriteAt(bufferContent[:bundleChunk.size], fileOffset)
	if errWriteChunk != nil {
		return errors.New("Erreur lors de l'écriture du chunk")
	}
	return nil
}

func downloadBundleFile(downloadUrl string, destinationFile string) error {
	res, err := http.Get(downloadUrl)
	if err != nil {
		return errors.New("Erreur de lien de telechargement d'un bundle " + err.Error())
	}
	defer res.Body.Close()

	file, errOpenFile := os.OpenFile(destinationFile, os.O_RDWR|os.O_CREATE, os.ModePerm)
	if errOpenFile != nil {
		return errors.New("Erreur lors de l'ouverture d'un fichier bundle" + errOpenFile.Error())
	}
	defer file.Close()

	_, errCopyToFile := io.Copy(file, res.Body)
	if errCopyToFile != nil {
		return errors.New("Erreur lors de la copie du contenu vers le fichier" + errCopyToFile.Error())
	}
	return nil
}
