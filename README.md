# Office Check-in Backend Service

Backend Service für die cronos Implementation des E.ON Office Check-in.
Dieser Service wurde entwickelt, um unabhängiger von Firebase zu sein und gleichzeitig eine große Menge neuer Features zu implementieren.

Die Benutzerauthentifizierung an sich läuft weiterhin über Firebase. Der Service prägt die AuthClient schnittstelle aus, um die von Firebase erstellten Tokens zu verifizieren und zu verarbeiten.
## Aufgaben

* Kontrollieren, ob Benutzer ein Admin ist
* Anlegen von Buchungen
* Löschen von Buchungen
* Einsehen von Buchungen
* Einsehen von Bereichen
* COVID-Backtracing
* Gesamtübersichtserstellung
* Datenexport
* Auto-Delete der alten DB-Einträge

## API

Die API ist unter dem Stammpfad ``/v1`` erreichbar.
Die _Definition der Routen_ findet sich in der Datei ``main.go``.

## Einrichtung

Das Backend verwendet eine MongoDB als Datenbank zum speichern der Daten. In der cronos Unternehmensberatung haben wir dazu auf die Cloud-Variante von MongoDB gesetzt. Sie haben aber auch die Möglichkeit, eine eigene MongoDB-Instanz zu hosten.
Für die Benutzerverwaltung wird Firebase verwendet. Genauere Informationen zur Einrichtung des Firebase Projektes finden Sie im Frontend-Projekt.
Im Backend wird eine Firebase-Konfigurationsdatei benötigt. Unter https://firebase.google.com/docs/admin/setup finden Sie eine Anleitung zum erstellen einer Firebase JSON Datei.
Diese Datei muss im Root-Verzeichnis des Projektes als ``firebase.json`` gespeichert werden.

Um das Backend zu deployen gibt es verschiedene Möglichkeiten.
Zunächst müssen Sie die ``config.yaml``-Konfigurationsdatei erstellen. Eine beispielhafte Datei finden Sie unter ``config.example.yaml``.
Kopieren Sie diese Datei und passen Sie die Einstellungen an.

### Einrichtung ohne Docker

Wenn Sie das Backend ohne Docker deployen möchten, installieren Sie go auf dem Host-Betriebssystem. Weitere Informationen finden Sie hier: https://golang.org/doc/install
Nach der Einrichtung von go, führen Sie folgende Schritte aus:

```
git clone https://github.com/cronosgmbh/office-checkin-backend.git $GOPATH/src/github.com/cronosgmbh/office-checkin-backend
cd $GOPATH/src/github.com/cronosgmbh/office-checkin-backend
go get -u ./...
go build .
office-checkin-backend
```

Der Service ist nun unter dem Port 3000 verfügbar.

### Einrichtung mit Docker

Wenn Sie das Deployment über Docker durchführen wollen, benötigen Sie einen Server mit installiertem Docker.

Wenn Sie die Applikation in Azure durchführen wollen, müssen Sie einen neuen App-Service anlegen.

Bauen Sie anschließend die Images lokal oder auf einem virtuellen server und pushen Sie die Abbilder in eine Container-Registry.

Öffnen Sie nun den App Service im Azure Portal. Klicken Sie links auf "Deployment Center". Klicken Sie auf Settings und wählen Sie als Container Type "Docker Compose (preview) aus"

```
version: '3.1'
services:
  frontend:
    image: containerregistry.example.com/ccheckin/oci-frontend:production
    container_name: oci_frontend
    hostname: frontend
    networks:
      - traefik-network

  backend:
    container_name: oci_backend
    image: containerregistry.example.com/ccheckin/oci-backend:production
    hostname: backend
    networks:
      - traefik-network

  reverse-proxy:
    image: containerregistry.example.com/ccheckin/oci-reverse-proxy:production
    container_name: reverse_nginx
    hostname: reverse-nginx
    networks:
      - traefik-network
    ports:
      - 80:80

networks:
  traefik-network:
    name: traefik-network
    external: false
```

Hinweis: Bevor Sie das Deployment durchführen, müssen Sie noch entsprechende Images für das Frontend und den Reverse Proxy.

Anleitungen dafür finden sie in den entsprechenden Repositories.