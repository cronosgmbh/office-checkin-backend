# Office Check-in Backend Service

Backend Service für die cronos Implementation des E.ON Office Check-in.
Dieser Service wurde entwickelt, um unabhängiger von Firebase zu sein und gleichzeitig eine große Menge neuer Funktionalitäten zu implementieren.

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
## Lokale Entwicklung
Folgende Umgebungsvariablen werden verwendet:

* ``CRONOS_ENV``: Umgebung des Deployments (development | staging | production)
* ``CRONOS_MONGO_HOST``: Hostname des MongoDB Servers
* ``CRONOS_MONGO_DB``: Datenbank in der MongoDB
* ``CRONOS_MONGO_USER``: Benutzername für die MongoDB-Authentifizierung
* ``CRONOS_MONGO_PASSWORD``: Passwort für die Mongo-DB-Authentifizierung

Außerdem muss ein Dienstaccount für ein Firebase-Projekt eingerichtet und im Stammverzeichnis in eine firebase.json gelegt werden. 
