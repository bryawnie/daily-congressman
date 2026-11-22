# API

Servidor HTTP de solo lectura que entrega el congresista del día y los días anteriores. Lee únicamente la base SQLite que llenan los [jobs](../jobs): no llama a ninguna fuente externa, no necesita navegador y no escribe nada.

- Código: [`api.go`](api.go) (rutas, CORS, errores) y [`day.go`](day.go) (armado de cada día y tipos de la respuesta).
- Binario: [`cmd/api`](../../cmd/api/main.go).
- Tests: [`api_test.go`](api_test.go), con `go test ./internal/api`.

## Configuración

| Variable | Por defecto | Qué hace |
|---|---|---|
| `API_ADDR` | `:8080` | Dirección donde escucha |
| `API_CORS_ORIGINS` | `*` | Orígenes permitidos, separados por coma. `*` permite cualquiera; vacío no envía cabeceras CORS |
| `DB_PATH` | `data/congressmen.db` | Base SQLite (la misma que usan los jobs) |

Al iniciar aplica las migraciones pendientes (son idempotentes), de modo que puede arrancar antes que los jobs. Se apaga con gracia ante `SIGINT`/`SIGTERM`.

## Rutas

| Método y ruta | Respuesta |
|---|---|
| `GET /days?limit=8` | `{"days": [Day, ...]}`, el más reciente primero |
| `GET /days/{YYYY-MM-DD}` | Un [`Day`](#day) |
| `GET /health` | `{"status": "ok"}` |

Cualquier otro método da `405`. `OPTIONS` responde `204` (preflight CORS).

### `GET /days`

Parámetro `limit`: entero de 1 a 31, por defecto 8 (hoy y los 7 previos). Devuelve solo los días ya sorteados; con una base vacía responde `{"days": []}` con `200`. Si un día no se puede armar, falla toda la petición con `500` en lugar de entregar una lista más corta sin avisar.

### `GET /days/{fecha}`

La fecha debe ser `YYYY-MM-DD` (día de Chile). `404` si ese día no se sorteó.

### `GET /health`

`200` si la base responde, `503` si no.

## Errores

Todos tienen el mismo formato:

```json
{"error": "date must be YYYY-MM-DD"}
```

| Código | Cuándo |
|---|---|
| `400` | `limit` fuera de rango o no numérico; fecha mal formada |
| `404` | No hay congresista elegido para esa fecha |
| `405` | Método distinto de `GET` y `OPTIONS` |
| `500` | Error al leer la base (el detalle va al log, no a la respuesta) |
| `503` | `/health` no pudo consultar la base |

## Cabeceras

- `Content-Type: application/json; charset=utf-8`.
- `Cache-Control: public, max-age=300` en `/days` y `/days/{fecha}` (los datos cambian pocas veces al día).
- CORS: si el `Origin` de la petición está permitido, `Access-Control-Allow-Origin` (con `Vary: Origin` salvo en `*`), `Access-Control-Allow-Methods: GET, OPTIONS` y `Access-Control-Max-Age: 86400`. Un origen no permitido no recibe cabeceras CORS.

## Formato de `Day`

Los nombres de campo van en camelCase. Ejemplo (recortado):

```json
{
  "date": "2026-11-21",
  "beacon": { "chainId": "…", "pulseId": "…", "randOut": "…" },
  "congressman": {
    "id": 12,
    "chamber": "senate",
    "externalId": 1517,
    "name": "Ignacio",
    "firstSurname": "Urrutia",
    "secondSurname": "Bonilla",
    "picture": "https://…",
    "gender": "M",
    "birthdate": "1970-01-31",
    "profession": "Abogado",
    "party": "Partido …",
    "partyAlias": "…",
    "twitterUsername": "…",
    "instagramUsername": "…",
    "constituency": { "kind": "region", "name": "Región de …", "detail": "" },
    "periods": [{ "from": 2022, "to": 2030 }]
  },
  "attendance": { "attended": 80, "justifiedAbsent": 5, "unjustifiedAbsent": 2, "total": 87 },
  "activity": { "all": 3, "byStatus": { "En tramitación": 2, "Publicado": 1 } },
  "votings": [
    {
      "votingId": 123,
      "votingDate": "2026-11-18",
      "bulletinNumber": "18442-07",
      "documentTitle": "…",
      "articleText": "…",
      "voted": "Sí",
      "totalApproved": 30,
      "totalRejected": 15,
      "totalAbstention": 2,
      "quorum": "Mayoría simple",
      "result": "Aprobado"
    }
  ],
  "expenses": [
    {
      "year": 2026,
      "month": "octubre",
      "total": 5200000,
      "detail": [
        {
          "type": "Operacional",
          "amount": 5200000,
          "stats": {
            "global": { "_type": "all", "count": 50, "ranking": 4, "avg": 4800000, "min": 1200000, "max": 9100000 },
            "local":  { "_type": "region", "count": 3, "ranking": 1, "avg": 4100000, "min": 3000000, "max": 5200000 }
          },
          "expenses": [
            {
              "subtype": "Telefonía",
              "amount": 300000,
              "stats": {
                "global": { "_type": "all", "count": 50, "avg": 250000, "min": 0, "max": 900000 },
                "local":  { "_type": "region", "count": 3, "avg": 280000, "min": 200000, "max": 300000 }
              }
            }
          ]
        }
      ]
    }
  ],
  "pending": []
}
```

### Campos

**`date`**: día sorteado, `YYYY-MM-DD` (Chile).

**`beacon`**: de dónde salió el sorteo, para poder verificarlo. `chainId` y `pulseId` identifican el pulso del Faro (Beacon) de las 00:00 y `randOut` es su valor aleatorio. Con ellos, la lista de miembros y el algoritmo documentado cualquiera puede repetir el sorteo (`chooser --date`).

**`congressman`**: perfil. `chamber` es `deputies` o `senate`. `id` es el identificador interno y `externalId` el de la cámara de origen; solo el par `(chamber, externalId)` es único entre fuentes. `constituency.kind` es `district`, `region` o `""` si no se conoce. `periods` son los períodos parlamentarios, el más reciente primero. Los textos que la fuente no entrega vienen como `""`.

**`attendance`**: `null` si aún no se extrajo. `attended` son las asistencias; `total` las sesiones consideradas.

**`activity`**: `null` si aún no se extrajo. `all` es el total de proyectos y `byStatus` el conteo por estado tal como lo publica la fuente (los estados varían entre cámaras, por eso es un mapa).

**`votings`**: las últimas votaciones, la más reciente primero. `voted` es la opción del congresista. `quorum` es la regla exigida y `result` es `Aprobado` o `Rechazado`; el Senado no publica el resultado, así que se calcula desde el quórum, y puede venir `""` si no se pudo calcular.

**`expenses`**: hasta 6 meses con gasto operacional registrado, el más reciente primero. `month` está en español minúscula. `total` suma personal de apoyo (si la cámara lo informa) y gasto operacional. Cada elemento de `detail` es un tipo de gasto:

| Campo | Significado |
|---|---|
| `type` | `Operacional` o `Personal de Apoyo` (este último solo en las cámaras que lo informan, y va primero) |
| `amount` | Monto del mes en pesos. En `Operacional` es la suma de sus categorías; en `Personal de Apoyo` es `null` si el congresista no tiene registro ese mes |
| `supportStaff` | Personas contratadas (solo `Personal de Apoyo`, ausente en los demás) |
| `stats` | Comparación con otros congresistas (ver abajo) |
| `expenses` | Solo en `Operacional`: las categorías, cada una con `subtype`, `amount` y su propio `stats` |

**`stats`** (en cada tipo de gasto y en cada categoría) agrupa dos comparaciones del mismo mes:

```json
"stats": { "global": { ... }, "local": { ... } }
```

- `global`: todos los congresistas de **su misma cámara**. Siempre presente.
- `local`: solo quienes comparten su circunscripción, es decir el mismo distrito entre diputados o la misma región entre senadores (se compara `constituency.kind` y `constituency.name`). Es `null` cuando no se conoce la circunscripción del congresista, porque comparar "con los que no tienen circunscripción" no significaría nada.

Ambos tienen los mismos campos, e incluyen al propio congresista:

| Campo | Significado |
|---|---|
| `_type` | Según qué se agrupó: `all` en `global` (toda la cámara); en `local`, `district` (diputados del mismo distrito) o `region` (senadores de la misma región), el `constituency.kind` del congresista |
| `count` | Cuántos congresistas del grupo tienen registro ese mes |
| `ranking` | Posición en el grupo (1 = quien más gastó), de `count`. Solo en los totales de `Operacional` y `Personal de Apoyo`; las categorías no lo traen. Ausente si el congresista no tiene registro ese mes |
| `avg`, `min`, `max` | Promedio (redondeado), mínimo y máximo del grupo. `null` si nadie tiene registro |

Las comparaciones usan solo a quienes ya están en la base, así que mejoran a medida que los jobs mensuales sincronizan a todos.

**`pending`**: secciones que aún no se han extraído. Valores posibles: `attendance`, `votings`, `activity`, `expenses`. Siempre es un arreglo (vacío si no falta nada).

## Pendiente frente a cero

La API distingue "todavía no hay datos" de "no hay registros":

| Situación | Cómo se ve |
|---|---|
| Nunca se extrajo | `null` (`attendance`, `activity`) o `[]` (`votings`, `expenses`), y la sección aparece en `pending` |
| Se extrajo y no hay registros | `activity` con `all: 0`, `votings` o `expenses` como `[]`, y la sección **no** está en `pending` |

Para saberlo consulta el registro de trabajos completados (`job_state`): `refresh` para votaciones y proyectos, `expenses` para gastos. La asistencia se considera pendiente mientras no exista su fila. Los arreglos nunca son `null`, así el cliente puede recorrerlos sin comprobar.

Un congresista elegido siempre tiene `congressman` completo: si es el primer día de una base nueva, el job lee su perfil antes de guardar la elección.
