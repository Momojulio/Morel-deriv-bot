# Image de base Python
FROM python:3.10

# Dossier de travail
WORKDIR /app

# Copier les dépendances et les installer
COPY requirements.txt .
RUN pip install --no-cache-dir -r requirements.txt

# Copier le reste du code
COPY . .

# Port à exposer pour Render
EXPOSE 8080

# Commande de démarrage
CMD ["python", "main.py"]
