<h1>
  <a href="?">#DiputadxDelDia</a>
</h1>

> **🚧 Proyecto en reestructuración.**
> La rama `restructure-go` reescribe el proyecto en Go, por componentes. Ya funcionan el sorteo (*chooser*), la extracción de datos (*retrieval*) y los trabajos programados (*jobs*), por ahora **solo para el Senado**. La API y la webapp vienen después, y los diputados dependen de poder leer los datos de la Cámara (ver [Estado](#estado)).
> El código anterior (Python + Next.js) se conserva en [`legacy/`](legacy).

**#DiputadxDelDia** elige cada día a un(a) congresista de Chile usando el valor aleatorio del [Faro de Aleatoriedad de la Universidad de Chile](https://random.uchile.cl) a las 00:00 (hora de Chile). Una vez elegido, se reúne su información:

* Edad, profesión (si aplica), partido, y región o circunscripción que representa.
* Porcentaje de asistencia.
* Gastos (operacionales y, donde existan, oficinas y personal de apoyo).
* Actividad parlamentaria (proyectos que ha presentado).
* Sus votaciones más recientes.

## Componentes

| Componente | Dónde | Qué hace |
|---|---|---|
| Retrieval | `internal/source`, `internal/retrieval`, `cmd/retrieval` | Lee los datos de cada cámara y los guarda en SQLite |
| Chooser | `internal/chooser`, `internal/beacon`, `cmd/chooser` | Sorteo del día a partir del pulso del Faro |
| Jobs | `internal/jobs`, `cmd/jobs` | Programa la sincronización y el sorteo |
| API | `cmd/api` | Pendiente |
| Webapp | `webapp/` | Pendiente |

Cada cámara es una `Source` (`internal/congress/source.go`) registrada en un `Registry`; agregar una cámara es implementar esa interfaz. Los datos van en SQLite (`DB_PATH`, por defecto `data/congressmen.db`).

## Estado

* **Senado:** perfil, asistencia, votaciones, proyectos y gastos operacionales, desde la API pública de `web-back.senado.cl`. El Senado no publica el resultado de cada votación, así que se **calcula** a partir del quórum (se guarda junto al resultado).
* **Diputados:** solo la lista vigente (se usa en el sorteo). El resto depende de `camara.cl`, protegida por Cloudflare, y no se ha resuelto. No se debe activar `deputies` en `CHOOSER_CHAMBERS` hasta que su perfil funcione.

## Cómo se elige al congresista del día

El algoritmo es público y reproducible:

1. Se toma el pulso del Faro generado a las 00:00 de Chile (`America/Santiago`) del día.
2. Se ordenan todos los miembros vigentes de las cámaras activas por `(cámara, id externo)`.
3. Se lee el valor de salida (`outputValue`) del pulso como un entero hexadecimal grande `R`.
4. Se elige al miembro de índice `R mod N`, con `N` el número de miembros.

La elección se guarda al hacerla y nunca se recalcula. Se asume que la lista de miembros no cambia entre el sorteo y una verificación posterior.

### Verificar un día

Cualquiera puede repetir el sorteo, porque todas sus entradas son públicas. Con Go instalado:

```bash
go run ./cmd/chooser --date 2026-11-22
```

Imprime el día, la cámara, el congresista y la dirección del pulso del Faro. Usa las cámaras de `CHOOSER_CHAMBERS`.

## Trabajos programados

`jobs` hace cuatro cosas, todas idempotentes y registradas en la base:

| Trabajo | Frecuencia | Qué hace |
|---|---|---|
| `profiles` | 1 vez al mes | Guarda el perfil de todos los congresistas |
| `expenses` | 1 vez al mes | Guarda los gastos de todos, para compararlos |
| `choose` | 1 vez al día | Sortea cuando el pulso de las 00:00 está disponible |
| `refresh` | 1 vez al día | Actualiza asistencia, votaciones y proyectos del elegido |

```bash
go run ./cmd/jobs serve      # proceso permanente: revisa qué está pendiente cada JOBS_INTERVAL
go run ./cmd/jobs run        # una sola pasada (para un cron externo)
go run ./cmd/jobs daily      # sortea y refresca hoy (-date YYYY-MM-DD, -force refresca de nuevo)
go run ./cmd/jobs profiles   # sincronización mensual de perfiles (-force)
go run ./cmd/jobs expenses   # sincronización mensual de gastos (-force)
```

Como el trabajo hecho en cada período (día o mes de Chile) queda registrado, se puede invocar tantas veces como se quiera: lo ya hecho se omite, lo que falló se reintenta, y con todo al día no se hace ninguna petición. Un día ya sorteado nunca se vuelve a sortear.

Otros comandos útiles:

```bash
go run ./cmd/retrieval migrate                 # crea la base y aplica migraciones
go run ./cmd/retrieval fetch senate 1517       # muestra los datos de un senador, sin guardar
go run ./cmd/retrieval update senate 1517      # los guarda
```

## Configuración

Variables de entorno (ver [`.env.example`](.env.example)):

| Variable | Por defecto | Para qué |
|---|---|---|
| `CHOOSER_CHAMBERS` | `senate` | Cámaras que participan: `senate`, `deputies` o ambas separadas por coma |
| `DB_PATH` | `data/congressmen.db` | Archivo SQLite |
| `JOBS_INTERVAL` | `15m` | Cada cuánto `jobs serve` revisa qué está pendiente |
| `JOBS_PAUSE` | `500ms` | Espera entre congresistas en las sincronizaciones mensuales |

## Desarrollo

Requiere Go (ver `go.mod`).

```bash
make build   # compila los binarios en bin/
make test    # go test ./...
make vet
```

Las pruebas contra los servicios reales del Senado están aparte: `go test -tags live -v ./internal/source/senate`.

## Despliegue (VPS con Docker)

```bash
cp .env.example .env     # ajustar CHOOSER_CHAMBERS, etc.
docker compose up --build -d
```

Levanta `jobs serve` con la base en el volumen `data`. La imagen (`deploy/Dockerfile`) se construye por binario con `--build-arg BINARY=jobs|api`; la API se agregará con el mismo volumen cuando exista.

## Legacy

El proyecto anterior (Python + Next.js) está en [`legacy/`](legacy); sus instrucciones, en [`legacy/README.md`](legacy/README.md).
