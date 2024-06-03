# Go Server for Utleieadministrasjon

Denne serveren er opprettet for å administrere utleie av utstyr, sjekke utleiestatus, og sende ut e-postmeldinger for utleid utstyr som ikke er innlevert. Den benytter PocketBase for databasestyring og Echo som web-rammeverk.

## Installasjon

For å kjøre denne serveren må du ha Go installert på maskinen. Du kan laste ned og installere Go fra [https://golang.org/dl/](https://golang.org/dl/).

1. Klone dette prosjektet til din lokale maskin:
   ```bash
   git clone https://github.com/dittGithubRepo/dittProsjektNavn.git
   cd dittProsjektNavn
   ```

2. Installer nødvendige Go-pakker:
   ```bash
   go mod tidy
   ```

## Miljøkonfigurasjon

Konfigurer serveren ved å opprette en `.env`-fil i rotmappen til prosjektet. Denne filen skal inneholde følgende nøkler:

```
CRON_SCHEDULE="0 18 * * *"
E_MAIL="mottakerens_email@example.com"
```

- `CRON_SCHEDULE`: Tidspunktet for når e-poster skal sendes daglig (her kl. 18:00).
- `E_MAIL`: E-postadressen til mottakeren av daglige rapporter.

## Kjøre serveren

For å starte serveren, bruk følgende kommando:

```bash
go run main.go
```

Du kan også spesifisere hvilken HTTP-port serveren skal lytte på bruk kommandolinjeflagget `--http`:

```bash
go run main.go --http :8080
```

### API-endepunkter

Serveren tilbyr følgende API-endepunkter:

- **POST /create**: Oppretter et nytt utleieelement i `inventory`-samlingen.
- **GET /reserve/:id**: Markerer et element som reservert basert på `id`.
- **GET /inactive/:id**: Markerer et element som inaktivt/uaktuell basert på `id`.
- **GET /deliver/:id**: Registrerer innlevering av utleid utstyr og oppdaterer relevante datarekorder.

## Cron-job

En cron-job kjøres daglig kl. 18:00 for å sjekke for utleid utstyr som ikke er innlevert og sender e-postmeldinger til de relevante personene.

## Sikkerhet

Vær oppmerksom på sikkerhetsinnstillinger, spesielt rundt eksponering av sensitive API-endepunkter og miljøvariabler.

## Feilsøking

For å diagnostisere eventuelle problemer, sjekk loggene som produseres av serveren.
