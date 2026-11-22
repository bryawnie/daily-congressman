# #DiputadxDelDia (legacy: Python + Next.js)

> Este es el proyecto anterior, conservado hasta que cada pieza nueva en Go dé los mismos resultados. La versión actual está en el [README de la raíz](../README.md).

Lo que sigue describe el proyecto anterior, que obtenía los datos de [www.camara.cl](https://www.camara.cl) para diputados. Las rutas son relativas a esta carpeta (`backend/`, `frontend/`); desde la raíz del repositorio son `legacy/backend/` y `legacy/frontend/`.

## Verificar la elección con el script legacy

Dentro del presente repositorio se dispone de un script de verificación, el cual recrea el proceso que Random UChile lleva a cabo todos los días a las 00:00 horas (Chile Continental) para elegir al (a la) #DiputadxDelDia.

### Requisitos

* Python 3.0+
* Instalar requerimientos: `$ pip3 install -r backend/requirements.txt`

_**Nota:** Sugerimos utilizar un ambiente virtual._

### Uso
El _script_ se encuentra dentro de la carpeta `backend`, y es llamado `updater.py`. El _script_ recibe los siguientes parámetros:
```
$ python updater.py --verify [opcion] [valor]

Opciones:
-d [date]   Establece la fecha de generación del valor aleatorio a utilizar (formato DD-MM-AAAA).
-t [time]   Establece la hora de generación del valor aleatorio a utilizar (formato HH:MM).
-e [epoch]  Establece la fecha y hora de generación del valor aleatorio a utilizar en formato epoch en huso horario GMT (https://www.epochconverter.com/).
```

### Ejemplo

Si quiere verificar el #DiputadxDelDia del **20 de Enero de 2023** puede ejecutar uno de los siguientes dos comandos:

```
$ python3 updater.py --verify -d 2023-01-30 -t 00:00
```
```
$ python3 updater.py --verify -e 1675047600000
```

## Levantar el proyecto legacy
La aplicación cuenta con un *backend* en Flask, el cual se encarga de servir los datos de lxs diputadxs. Al mismo tiempo, cuenta con un *frontend* en Next.js el cual consume los datos y los muestra de forma más amigable. 

## Con Docker
El `docker-compose.yml` de la raíz ahora levanta los jobs en Go. Para el stack legacy hay que construir sus imágenes a mano desde `backend/` y `frontend/`, que tienen cada una su `Dockerfile`.

## Sin Docker (legacy)
**Requisitos**: Python3.6+, npm.

#### Backend
Primero será necesario levantar el backend, para ello seguimos los siguientes pasos:
```bash
# Cambio de directorio
$ cd backend

# Instalación de requisitos
$ pip3 install -r requirements.txt

# Inicializar base de datos
$ python3 updater.py -c -i

# Obtener data del último diputado.
$ python3 updater.py -r

# Servir el Backend
$ python3 webapp.py
```
#### Frontend
Para levantar el frontend, habrá que usar *node package updater*:
```
# Cambio de directorio
$ cd frontend

# Instalación de paquetes
$ npm install

# Servir Frontend
$ npm run serve
```
