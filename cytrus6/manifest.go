package cytrus6

import (
	"cytrusdownloader/cytrus6/flatbuffer"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"strconv"
)

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

func downloadManifest(game string, release string, platform string, version string) ([]byte, error) {
	// télécharge le fichier de manifest
	res, err := http.Get(fmt.Sprintf("https://cytrus.cdn.ankama.com/%s/releases/%s/%s/%s.manifest", game, release, platform, version))
	if err != nil {
		return []byte{}, errors.New("Erreur lors de la requete du téléchargement du fichier manifest, erreur:" + strconv.FormatInt(int64(res.StatusCode), 10))
	}

	defer res.Body.Close()
	data, errReadBody := io.ReadAll(res.Body)
	if errReadBody != nil {
		return []byte{}, errors.New("Erreur lors de la lecture du corps de la requête")
	}
	return data, nil
}

func extractManifestFromFileData(data []byte) Manifest {
	//func convertManifestFileToManifest(data []byte) Manifest {
	// Parse le fichier manifest

	manifestExtracted := Manifest{}

	manifest := flatbuffer.GetRootAsManifest(data, 0)

	for i := range manifest.FragmentsLength() {
		// parse un fragment
		fragment := &flatbuffer.Fragment{}
		if success := manifest.Fragments(fragment, i); !success {
			log.Panic("Erreur lors du parsing du fichier manifest")
		}

		fragmentExtracted := Fragment{name: string(fragment.Name())}

		for j := range fragment.FilesLength() {
			// extrait les fichier
			file := &flatbuffer.File{}
			if success := fragment.Files(file, j); !success {
				log.Panic("Erreur lors du parsing du fichier manifest")
			}
			fileExtracted := File{name: string(file.Name()), size: file.Size(), executable: file.Executable(), symlink: string(file.Symlink()), hash: extractHash(file)}

			for k := range file.ChunksLength() {
				chunk := &flatbuffer.Chunk{}
				file.Chunks(chunk, k)
				fileExtracted.chunks = append(fileExtracted.chunks, Chunk{hash: extractHash(chunk), size: chunk.Size(), offset: chunk.Offset()})
			}
			fragmentExtracted.files = append(fragmentExtracted.files, fileExtracted)
		}

		for j := range fragment.BundlesLength() {
			// parse les fichiers bundle
			bundle := &flatbuffer.Bundle{}
			fragment.Bundles(bundle, j)
			bundleExtracted := Bundle{hash: extractHash(bundle)}

			for k := range bundle.ChunksLength() {
				chunk := &flatbuffer.Chunk{}
				bundle.Chunks(chunk, k)
				bundleExtracted.chunks = append(bundleExtracted.chunks, Chunk{hash: extractHash(chunk), size: chunk.Size(), offset: chunk.Offset()})
			}
			fragmentExtracted.bundles = append(fragmentExtracted.bundles, bundleExtracted)
		}
		manifestExtracted.fragments = append(manifestExtracted.fragments, fragmentExtracted)
	}

	return manifestExtracted
}

type HashObject interface {
	Hash(j int) byte
	HashLength() int
}

func extractHash(obj HashObject) string {
	hash := ""
	for i := range obj.HashLength() {
		//hash += string(strconv.FormatInt(int64(obj.Hash(i)), 16))
		hash += fmt.Sprintf("%02s", strconv.FormatInt(int64(obj.Hash(i)), 16))
	}
	return hash
}
