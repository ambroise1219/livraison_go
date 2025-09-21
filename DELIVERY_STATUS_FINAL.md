# 📋 STATUTS DÉFINITIFS PAR TYPE DE LIVRAISON

## 🔵 **LIVRAISON STANDARD**
```
1. PENDING                  → Commande créée
2. ASSIGNED                 → Livreur assigné
3. ARRIVED_AT_PICKUP        → Livreur arrivé chez expéditeur
4. PICKED_UP                → Colis récupéré
5. IN_TRANSIT               → En route vers destinataire  
6. ARRIVED_AT_DESTINATION   → Arrivé chez destinataire
7. DELIVERED                → ✅ Livraison terminée
8. CANCELLED                → ❌ Annulée (possible à tout moment avant DELIVERED)
```

## 🟠 **LIVRAISON EXPRESS** 
```
1. PENDING                  → Commande EXPRESS créée (priorité)
2. ASSIGNED                 → Livreur assigné (prioritaire)
3. ARRIVED_AT_PICKUP        → Livreur arrivé chez expéditeur
4. PICKED_UP                → Colis récupéré
5. IN_TRANSIT               → En route (rapide) vers destinataire
6. ARRIVED_AT_DESTINATION   → Arrivé chez destinataire
7. DELIVERED                → ✅ Livraison EXPRESS terminée
8. CANCELLED                → ❌ Annulée (possible à tout moment avant DELIVERED)
```

## 🟢 **LIVRAISON GROUPÉE**

### **Statuts Globaux (Livraison Parent):**
```
1. PENDING                  → Commande groupée créée
2. ASSIGNED                 → Livreur assigné
3. ARRIVED_AT_PICKUP        → Livreur arrivé chez vendeur (point unique)
4. PICKED_UP                → Tous les colis récupérés
5. DELIVERY_IN_PROGRESS     → Distribution des zones en cours
6. DELIVERED                → ✅ Toutes zones livrées
7. CANCELLED                → ❌ Annulée (possible à tout moment avant DELIVERED)
```

### **Statuts par Zone:**
```
Zone 1: (démarre directement après PICKED_UP global)
- IN_TRANSIT → ARRIVED_AT_DESTINATION → DELIVERED

Zone 2,3,4,5... (chaque zone démarre quand la précédente est DELIVERED)
- PENDING → IN_TRANSIT → ARRIVED_AT_DESTINATION → DELIVERED
```

## 🔴 **LIVRAISON DÉMÉNAGEMENT**
```
1. PENDING                     → Déménagement programmé
2. ASSIGNED                    → Équipe assignée
3. ASSIGNED_TO_HELPER          → Assistants assignés
4. HELPERS_CONFIRMED           → Équipe confirmée
5. ARRIVED_AT_PICKUP           → Équipe arrivée chez client
6. LOADING_IN_PROGRESS         → Chargement en cours
7. LOADING_COMPLETED           → Chargement terminé
8. IN_TRANSIT                  → Transport vers nouvelle adresse
9. ARRIVED_AT_DESTINATION      → Arrivé à destination
10. UNLOADING_IN_PROGRESS      → Déchargement en cours
11. UNLOADING_COMPLETED        → Déchargement terminé
12. DELIVERED                  → ✅ Déménagement terminé
13. CANCELLED                  → ❌ Annulé (possible à tout moment avant DELIVERED)
```

## 📊 **UNIFORMISATION DES TERMES:**

Pour une cohérence parfaite dans l'ensemble du système:
- ✅ `ARRIVED_AT_DESTINATION` est utilisé pour toutes les livraisons
- ❌ `ARRIVED_AT_DROPOFF` n'est plus utilisé

## 🔄 **LOGIQUE DE TRANSITION:**

### Pour Standard, Express et Déménagement:
- Transitions manuelles: le livreur marque lui-même chaque étape
- Ordre chronologique strict: impossible de sauter des étapes

### Pour Livraison Groupée:
- La livraison parent suit une progression manuelle jusqu'à `PICKED_UP`
- Transition automatique: `PICKED_UP` → `DELIVERY_IN_PROGRESS` dès que Zone 1 démarre
- Chaque zone suit son propre cycle (IN_TRANSIT → ARRIVED_AT_DESTINATION → DELIVERED)
- Quand toutes les zones sont `DELIVERED`, la livraison parent passe automatiquement à `DELIVERED`

## 🏭 **POUR IMPLÉMENTATION FUTURE:**

Statuts supplémentaires disponibles mais non utilisés dans le flux standard:
- `SORTING_IN_PROGRESS`, `SORTED` (pour centre logistique)
- `DISPATCH_IN_PROGRESS` (pour gestion d'entrepôt)
- `ZONE_ASSIGNED` (si réorganisation des zones devient dynamique)
- `EN_ROUTE` (générique, remplacé par le plus précis IN_TRANSIT)