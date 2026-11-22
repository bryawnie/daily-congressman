# Webapp

Sitio estático en SvelteKit que muestra el parlamentario del día y los 7 días anteriores. Se construye consultando la [API](../internal/api) y no necesita servidor en producción: el resultado queda en `build/`.

## Comandos

Requiere Node 22.17 o superior (`.nvmrc` fija 26.13.1, la LTS vigente).

| Comando | Qué hace |
|---|---|
| `npm install` | Instala las dependencias |
| `npm run dev` | Servidor de desarrollo (necesita la API corriendo) |
| `npm run check` | Revisa tipos y componentes |
| `npm run build` | Genera el sitio estático en `build/` |
| `npm run preview` | Sirve `build/` para revisarlo |

Con `API_URL` se elige de dónde se leen los datos al construir (por defecto `http://localhost:8080`). Si la API no responde, la construcción falla: nunca se publica un sitio con datos a medias.

## Cómo se arma

- `/` muestra el día más reciente y `/dias/AAAA-MM-DD` cada uno de los últimos 8 días. Todas las páginas se prerrenderizan, así que hay que volver a construir después de que el job sortee y refresque (una vez al día).
- Una sección que la API lista en `pending` muestra un aviso en lugar de datos. Una sección sin registros muestra que no hay, que es distinto.
- Si la API no tiene días, `/` avisa que aún no hay sorteo y no se dibuja el selector. Una URL inexistente muestra `+error.svelte` (`404.html` como respaldo del hosting).
- Todos los tipos viven en `src/lib/types.ts`, incluidas las props de cada componente.
- No hay librerías de gráficos: las barras, el gráfico mensual y las escalas de comparación son HTML y CSS.
- Tema claro y oscuro según el sistema, y diseño responsivo (se probó desde 390 px).

## Estructura

```
src/
├── env.ts                 # variables de entorno (API_URL)
├── routes/                # solo carga de datos y composición
└── lib/
    ├── types.ts           # todos los tipos
    ├── api/               # cliente de la API
    ├── server/            # acceso a la API desde las cargas (solo servidor)
    ├── format/            # fechas, pesos, porcentajes y nombres
    ├── expenses/          # reglas de comparación de gastos
    ├── votings/           # colores de votos y resultados
    ├── styles/            # tokens y estilos globales
    └── components/
        ├── ui/            # piezas genéricas: Badge, SectionCard, StackedBar, ...
        ├── layout/        # cabecera y pie
        ├── day/           # página del día, selector de días, verificación del sorteo
        ├── profile/       # ficha del parlamentario
        ├── attendance/  activity/  votings/  expenses/
```
