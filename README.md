# Cytrus-downloader
Un outil pour télécharger les jeux en provenance du cdn d'ankama
Il est compatible avec cytrus 5 et 6

## Télécharger et compiler l'outil:
- Avoir Go 1.23 ou supérieur d'installé
```
git clone https://github.com/loonaire/cytrus-downloader.git
cd cytrus-downloader
go build
```
Vous pouvez ensuite utiliser le fichier cytrus-downloader.exe

## Utilisation:
Télécharger la dernière version d'un jeu:
```
./cytrus-downloader.exe -game retro -platform windows -release main
```

## Remerciements

- https://github.com/nexepu/Nexytrus/ Pour la partie cytrus 5
- https://github.com/ledouxm/cytrus-v6
- https://github.com/AlpaGit/cytrus-downloader-v6 Pour le flatbuffer de cytrus 6


