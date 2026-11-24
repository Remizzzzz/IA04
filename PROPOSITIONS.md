# Propositions d'amélioration du projet

## 1. Simplification des unités

Le jeu conservera trois unités de combat :

- **Infantry** : soldat standard et polyvalent.
- **Sniper** : unité peu résistante, avec une vision et une portée d'attaque supérieures.
- **Tank** : unité résistante, avec des dégâts élevés et une portée d'attaque spécifique.

Je suggère de ne pas intégrer le Supply Truck dans un premier temps et d'utiliser des munitions illimitées pour toutes les unités.

La gestion du ravitaillement pourrait compliquer un peu la programmation. Étant donné que l'objectif principal du projet est la simulation et l'utilisation des goroutines, il serait préférable de se concentrer d'abord sur les unités de combat, leurs déplacements, leurs interactions et la simulation des combats.

L'ajout du Supply Truck pourrait être envisagé plus tard, en fonction de l'avancement du projet.

## 2. Caractéristiques des unités

Valeurs initiales proposées pour les PV :

  | Unité   | PV        |
  |-------- |-----------|
  |Infantry |          2|
  |Sniper   |          1|
  |Tank     |          5|



## 3. Portée et dégâts des attaques

Afin de différencier les unités et de simplifier le système de combat, chaque type d'unité disposera de dégâts fixes et d'une portée d'attaque spécifique.

**Caractéristiques proposées :**

| Unité | Dégâts par attaque | Portée d'attaque | Particularité |
|---|---:|---:|---|
| Infantry | 1 PV | 3 cases | Unité polyvalente |
| Sniper | 2 PV | 7 cases | Faible résistance, portée supérieure |
| Tank | 2 PV | 4 cases | Forte résistance et puissance offensive |

**Règles de combat :**

- Chaque attaque cible une seule unité ennemie.
- Les dégâts sont fixes : aucun tir aléatoire.
- Une unité ne peut attaquer qu'une cible située dans sa portée d'attaque.
- Une unité dont les PV atteignent 0 est éliminée.
- Le Tank ne dispose pas d'attaque de zone dans cette première version.

**Justification :**

Cette approche permet de différencier les unités selon leur rôle : l'Infantry constitue une unité polyvalente, le Sniper privilégie les attaques à longue distance et le Tank joue un rôle offensif et défensif grâce à sa résistance.

Les valeurs proposées pourront être ajustées après plusieurs simulations afin d'évaluer l'équilibre entre les unités et l'efficacité des différentes stratégies.

## 4. Résolution des tirs : tirs simultanés par tour

Toutes les unités vivantes au début de la phase de combat peuvent effectuer une attaque pendant cette phase.

Le déroulement est le suivant :

1. Identifier toutes les unités vivantes au début de la phase.
2. Chaque unité sélectionne une cible ennemie située dans sa portée d'attaque.
3. Calculer les dégâts de chaque attaque sans modifier immédiatement les points de vie des cibles.
4. Cumuler les dégâts reçus par chaque unité.
5. Appliquer simultanément tous les dégâts.
6. Éliminer les unités dont les points de vie sont inférieurs ou égaux à zéro.

Une unité éliminée pendant cette phase peut donc terminer son attaque si elle était encore vivante au début de celle-ci.

**Justification :**

Cette approche évite que l'ordre d'exécution des unités influence le résultat des combats. Elle garantit également un comportement cohérent et facilite l'implémentation ainsi que les simulations des différentes stratégies.

## 5. Conditions de victoire

Deux conditions de victoire :

-   **Élimination** : gagner en éliminant toutes les unités adverses.
-   **Capture d'objectif** : atteindre une position désignée sur la
    carte et la contrôler pendant 2 ou 3 tours consécutifs.

Prévoir également un nombre maximal de tours afin d'éviter les
simulations infinies. Si aucune équipe ne remplit une condition de
victoire à la fin, le résultat peut être déclaré nul.

## 6. Élimination du général

Je propose de supprimer le général du jeu afin de simplifier les règles et d'éviter d'ajouter une mécanique spécifique liée à sa mort.

Toutes les unités suivront directement la stratégie de leur équipe et pourront continuer à se coordonner entre elles, sans dépendre d'un chef.

Cela permettra de se concentrer davantage sur les stratégies collectives et le comportement des unités, sans ajouter de complexité inutile.

## 7. Simplification des stratégies

Je propose de commencer par implémenter les règles de base du jeu et de conserver les quatre stratégies existantes (S0, S1, S2 et S3).

La stratégie liée à la logistique et au ravitaillement sera définitivement exclue du projet, notamment en raison de la suppression du Supply Truck et de l'utilisation de munitions illimitées.

Les autres stratégies et mécaniques supplémentaires ne seront pas exclues définitivement. Elles pourront être ajoutées dans une deuxième étape, lorsque les fonctionnalités principales seront opérationnelles, en fonction de l'avancement du projet et du temps disponible.

Cette approche permettra de développer et de tester progressivement le jeu, sans rendre la première version trop complexe.

## 8. Clarification des noms de variables (suggestion)

Je propose de clarifier certains noms de variables afin de rendre le code plus facile à comprendre et à maintenir. L'objectif est de privilégier des noms simples et explicites, plutôt que des termes trop techniques ou trop spécifiques au vocabulaire militaire.

Quelques exemples de renommage :

- `Kind` → `UnitType`
- `allegiance` → `teamID`
- `Reachable`  → `ValidMoves` 


