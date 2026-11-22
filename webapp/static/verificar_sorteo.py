"""Repite el sorteo del día (hoy, solo Senado). Uso: python3 verificar_sorteo.py 2026-11-22"""
import json
import sys
from datetime import datetime
from urllib.request import Request, urlopen
from zoneinfo import ZoneInfo

FARO = "https://random.uchile.cl/beacon/2.1-beta"
SENADORES = "https://web-back.senado.cl/api/hemicycle?vigentes=1&camara=S&limit=100"

def get_json(url):
    with urlopen(Request(url, headers={"User-Agent": "verificar-sorteo"}), timeout=30) as r:
        return json.load(r)

dia = datetime.strptime(sys.argv[1], "%Y-%m-%d").replace(tzinfo=ZoneInfo("America/Santiago"))
pulso = get_json(f"{FARO}/pulse?timeGE={int(dia.timestamp()) * 1000}")["pulse"]
senadores = sorted(
    get_json(SENADORES)["data"]["parlamentarios"]["data"], key=lambda s: s["ID_PARLAMENTARIO"]
)

elegido = senadores[int(pulso["outputValue"], 16) % len(senadores)]
print(" ".join(f"{elegido['NOMBRE']} {elegido['APELLIDO_PATERNO']} {elegido['APELLIDO_MATERNO']}".split()))
