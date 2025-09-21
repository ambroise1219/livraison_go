# 📋 STATUTS DE LIVRAISONS PAR TYPE

## 🔵 **LIVRAISON STANDARD** (Livraison simple classique)
```
1. PENDING                  → Commande créée, en attente
2. ASSIGNED                 → Livreur assigné  
3. PICKED_UP                → Colis récupéré chez l'expéditeur
4. IN_TRANSIT               → En route vers la destination
5. ARRIVED_AT_DROPOFF       → Arrivé à l'adresse de livraison
6. DELIVERED                → ✅ Livraison terminée
7. CANCELLED                → ❌ Annulée (à tout moment avant DELIVERED)
```

## 🟠 **LIVRAISON EXPRESS** (Livraison rapide prioritaire)
```
1. PENDING                  → Commande EXPRESS créée - PRIORITÉ
2. ASSIGNED                 → Livreur assigné immédiatement
3. PICKED_UP                → Récupération prioritaire
4. IN_TRANSIT               → Transport EXPRESS direct  
5. ARRIVED_AT_DROPOFF       → Arrivé rapidement
6. DELIVERED                → ✅ EXPRESS livrée
7. CANCELLED                → ❌ Annulée (à tout moment avant DELIVERED)
```

## 🟢 **LIVRAISON GROUPÉE** (Plusieurs clients - UN vendeur)
### 🏢 **Statuts Globaux (Livraison Parent):**
```
1. PENDING                  → Commande groupée créée avec N zones
2. ASSIGNED                 → Livreur assigné pour toutes les zones
3. PICKED_UP                → TOUS les colis récupérés chez le vendeur (point unique)
4. DELIVERY_IN_PROGRESS     → Livraisons des zones en cours
5. DELIVERED                → ✅ TOUTES les zones terminées
6. CANCELLED                → ❌ Annulée
```

### 📍 **Statuts par Zone Individuelle:**
```
Zone 1: PICKUP_COMPLETED → IN_TRANSIT → ARRIVED_AT_DROPOFF → DELIVERED
Zone 2: PENDING → IN_TRANSIT → ARRIVED_AT_DROPOFF → DELIVERED  
Zone 3: PENDING → IN_TRANSIT → ARRIVED_AT_DROPOFF → DELIVERED
Zone N: PENDING → IN_TRANSIT → ARRIVED_AT_DROPOFF → DELIVERED

📌 LOGIQUE: Quand Zone N est DELIVERED → Zone N+1 passe de PENDING à IN_TRANSIT
```

## 🔴 **LIVRAISON DÉMÉNAGEMENT** (Transport mobilier + équipe)
```
1. PENDING                     → Déménagement programmé
2. ASSIGNED                    → Équipe assignée
3. ASSIGNED_TO_HELPER          → Assistants assignés
4. HELPERS_CONFIRMED           → Équipe confirmée (ex: 4 personnes)
5. ARRIVED_AT_PICKUP           → Équipe arrivée chez le client
6. LOADING_IN_PROGRESS         → Chargement du mobilier en cours
7. LOADING_COMPLETED           → Tout chargé dans le camion
8. IN_TRANSIT                  → Transport vers la nouvelle adresse
9. ARRIVED_AT_DESTINATION      → Arrivé à destination
10. UNLOADING_IN_PROGRESS      → Déchargement + assemblage en cours
11. UNLOADING_COMPLETED        → Tout déchargé et assemblé
12. DELIVERED                  → ✅ Déménagement terminé
13. CANCELLED                  → ❌ Annulé
```

## 🏭 **STATUTS SPÉCIAUX AVANCÉS** (Entrepôts/Tri)
```
SORTING_IN_PROGRESS         → Tri des colis en cours
SORTED                      → Colis triés
DISPATCH_IN_PROGRESS        → Distribution en cours
ZONE_ASSIGNED               → Zones géographiques assignées
PICKUP_IN_PROGRESS          → Collecte en cours
PICKUP_COMPLETED            → Collecte terminée
EN_ROUTE                    → En route (générique)
```

## 📊 **RÉSUMÉ PAR COMPLEXITÉ:**

| Type | Nombre d'étapes | Complexité |
|------|----------------|------------|
| **STANDARD** | 6-7 étapes | ⭐⭐ Simple |
| **EXPRESS** | 6-7 étapes | ⭐⭐ Simple mais rapide |
| **GROUPÉE** | 4 globaux + N zones | ⭐⭐⭐⭐ Complexe multi-zones |
| **DÉMÉNAGEMENT** | 12-13 étapes | ⭐⭐⭐⭐⭐ Très complexe |

## 🎯 **LOGIQUE GROUPÉE DÉTAILLÉE:**

### Phase 1: Préparation (Vendeur → Livreur)
```
PENDING → ASSIGNED → PICKED_UP (point unique chez vendeur)
```

### Phase 2: Distribution Séquentielle
```
🎯 Zone 1: PICKUP_COMPLETED → IN_TRANSIT → ARRIVED_AT_DROPOFF → DELIVERED
⏳ Zone 2: PENDING (attente)
⏳ Zone 3: PENDING (attente)
⏳ Zone 4: PENDING (attente)
⏳ Zone 5: PENDING (attente)

🎯 Zone 1 DELIVERED → Zone 2: PENDING → IN_TRANSIT → ARRIVED_AT_DROPOFF → DELIVERED
⏳ Zone 3,4,5: PENDING (attente)

🎯 Zone 2 DELIVERED → Zone 3: PENDING → IN_TRANSIT → ARRIVED_AT_DROPOFF → DELIVERED
⏳ Zone 4,5: PENDING (attente)

... et ainsi de suite jusqu'à la dernière zone
```

### Phase 3: Finalisation
```
Toutes zones DELIVERED → Livraison globale = DELIVERED
```

## 🔄 **TRANSITIONS POSSIBLES:**
- **Vers CANCELLED**: Depuis n'importe quel statut avant DELIVERED
- **Linéaire**: La plupart des statuts suivent un ordre séquentiel
- **Parallèle**: Seulement pour GROUPÉE (zones multiples en parallèle conceptuel)
