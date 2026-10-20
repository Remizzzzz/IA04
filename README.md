# Coordination et concentration des forces : simulation multi-agent d'un combat tactique

## Descriptif

Deux équipes de X unités s'affrontent sur une carte 2D en grille comportant des obstacles qui bloquent les déplacements et les tirs. Chaque équipe compte X soldats (rapides, portée moyenne), Y tanks (lent, résistant, dégâts élevés) et Z camions (sans arme, ravitaille en munitions les alliés proches).

Chaque unité est un agent autonome exécuté dans sa propre **goroutine** : à chaque tick, elle reçoit sa perception locale (ennemis et alliés visibles, obstacles, messages), choisit une action (se déplacer, tirer, ravitailler, attendre) et l'envoie par **channel** à un contrôleur central. Celui-ci attend toutes les actions, résout les conflits dans un ordre déterministe et met à jour le monde, ce qui garantit des simulations reproductibles.

Les équipes suivent des stratégies de coordination croissante : **S0** individualiste (chacun vise l'ennemi le plus proche), **S1** partage des positions ennemies, **S2** concentration des tirs sur une cible commune, **S3** cohésion du groupe autour du tank.

Le backend est écrit en Go. Un client léger web ou lourd (ebiten) affiche la carte et le combat en temps réel (démarrer, pause, réinitialiser) ; un mode sans affichage enchaîne les simulations pour les expériences.

**Cadre théorique** : les lois de Lanchester (1916). Pour des combats en duels, la force d'une armée est proportionnelle à son effectif (loi linéaire) ; si les tirs se concentrent, elle l'est au carré de son effectif (loi carrée). La simulation permet de voir quelles règles individuelles font émerger l'une ou l'autre.

## Question

**À effectifs égaux, dans quelle mesure la coordination entre agents autonomes améliore-t-elle l'issue d'un combat, et jusqu'à quel désavantage numérique peut-elle compenser l'infériorité ?**

## Indicateurs clés

|Indicateur|Ce qu'il mesure|
|---|---|
|**Taux de victoire** (± intervalle de confiance, ≥ 200 runs par configuration)|Efficacité globale de chaque stratégie, en confrontation S0 à S3|
|**Survivants et PV restants** du vainqueur|Ampleur de la victoire, comparée aux prédictions de Lanchester|
|**Seuil de compensation** : effectif minimal (7 contre 6, 7 contre 5…) auquel l'équipe coordonnée gagne encore ≥ 50 %|Réponse directe à la seconde partie de la question|
|**Exposant de concentration p**, ajusté sur les courbes d'effectifs : dA/dt = −β·Bᵖ|0 ≈ duels (loi linéaire), 1 ≈ tirs concentrés (loi carrée)|
|**Tirs gaspillés** (sur une cible déjà condamnée) et **munitions consommées**|Coût de la coordination et rôle du ravitaillement|
|**Durée du combat**|Rapidité de la décision selon la stratégie|
