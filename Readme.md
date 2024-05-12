#Tickets de caisse

## Introduction
Ce projet est une application backend écrite en Golang pour gérer un système de traitement de tickets de caisse envoyés par les clients via un webhook HTTP, `POST /ticket`. L'objectif principal est de stocker ces tickets dans une base de données PostgreSQL tout en prenant en charge un volume potentiellement important de requêtes en utilisant un système de event broker RabbitMQ.

## Choix technologiques

### Golang
J'ai choisi Golang comme langage de programmation pour la gestion des requêtes HTTP et la logique métier de l'application. Golang est connu pour sa rapidité d'exécution et sa prise en charge native de la concurrence, ce qui en fait un choix idéal pour les applications nécessitant des performances élevées et la gestion de multiples requêtes simultanées.

### RabbitMQ 
Comme système de Event Broker pour prendre en charge un volume élevé de requêtes. Ce système permet de décharger immédiatement la charge de travail du traitement des requêtes HTTP et les traiter de manière asynchrone. Cela permet à notre application de maintenir des performances élevées même lorsqu'elle est confrontée à un trafic important.

## Utilisation
### Lancer l'application
J'ai fourni un fichier `docker-compose.yml` qui permet de lancer l'application et ses dépendances (PostgreSQL et RabbitMQ) en utilisant Docker. Pour lancer l'application, exécutez la commande suivante à la racine du projet:
```bash
docker-compose up
```
Cela lancera l'application sur le port 8080.

### Envoyer un ticket
Pour envoyer un ticket, vous pouvez utiliser un client HTTP tel que `curl` ou `Postman`. Voici un exemple de requête `POST` pour envoyer un ticket:
```bash
Webhook
POST http://localhost:8080/ticket

Payload:
Order: 123456
VAT: 3.10
Total: 16.90

product,product_id,price
Formule(s) midi,aZde,14.90
Café,IZ8z,2
```
### Base de données
Une base de données `tickets` Les deux tables `tickets`, `products` sont créées automatiquement lors du lancement de l'application:

Schema de la table `tickets`:
```sql
CREATE TABLE IF NOT EXISTS tickets (
    id SERIAL PRIMARY KEY,
    order_id VARCHAR(255),
    vat NUMERIC,
    total NUMERIC,
    valid BOOLEAN,
    ticket_text TEXT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
    );

id: Identifiant unique du ticket
order_id: Identifiant de commande du ticket
vat: TVA du ticket
total: Total du ticket
valid: Indique si le ticket est valide ou non
ticket_text: Texte brut du ticket
created_at: Date de création du ticket
```

Schema de la table `products`:
```sql
CREATE TABLE IF NOT EXISTS products (
    id SERIAL PRIMARY KEY,
    name VARCHAR(255),
    product_id VARCHAR(255) UNIQUE,
    price NUMERIC,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

id: Identifiant unique du produit
name: Nom du produit
product_id: Identifiant unique du produit
price: Prix du produit
created_at: Date de création du produit
```

### Test de charge
Pour tester la capacité de l'application à gérer un volume important de requêtes, j'ai utilisé l'outil `siege` pour envoyer un grand nombre de requêtes HTTP à l'application. 

Si vous n'avez pas `siege` installé sur votre machine, suivez ce guide
https://www.linode.com/docs/guides/load-testing-with-siege/

Pour lancer un test de charge, exécutez la commande suivante:
```bash
make test_load
```

Vous pouvez modifier les paramètres du test dans makefile en modifiant.

### Merci d'avance pour votre retour.


