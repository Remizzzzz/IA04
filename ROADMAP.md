# Tasks
## A1 : Finir la base de l'environnement de simulation
Implémenter les combats et les différentes unités à partir de la classe mère Unit : 
>Rappel : Unit code les mouvements de l'unité ainsi que sa mémoire et son allégeance. C'est la base de chaque unité : Un cerveau qui se rappelle de la map, de ses alliés et ennemis, et qui agit sur la base d'une stratégie. Pour l'instant, la stratégie ne concerne que les mouvements

Ce qu'il faut encore rajouter pour les units est leur PV, car elles sont toutes possiblement tuables. 

A partir d'Unit, il faut créer deux sous-classes : 
- Soldier : Ce sera la classe fille qui implémentera le combat, pour les unités combattantes
- Support : Ce sera la classe fille qui implémentera le support tactique (ravitaillement en armes, transports de soldats, etc...)

Une fois ces sous-classes créées et fonctionnelles, on aura la base de code suffisante pour commencer à implémenter nos  différentes stratégies et différentes façons de gagner.

### Règle de combat : 
- **Munitions** : Chaque tir consomme exactement 1 munition. Une unité soldier à court de munition se replie et/ou recherche un camion de ravitaillement.
- **Simultanéité** : actuellement, les déplacements d'un tour sont résolus dans l'ordre de création des unités, ce qui avantage les premières créées. Pour les tirs, il faudrait une résolution simultanée (tout le monde tire sur l'état du début de tour, puis on retire les morts), sinon l'ordre de création décide qui meurt avant d'avoir pu tirer.
- **Mort d'une unité** : la retirer de la grille, de `World.Units` et de `Army.Soldiers`, et arrêter sa goroutine.

### Règles du combat à trancher
> A fixer avant d'implémenter Soldier :
- **Tir** : Il faut que l'unité visée soit dans la vision (donc ligne de tir ouverte). Le tir touche-t-il toujours, ou avec une probabilité ? On mesurera la portée de manière euclidienne (à vol d'oiseau/en ligne direct)

- **Mort du général** : un nouveau général est-il désigné (et lequel), ou l'armée reste-t-elle désorganisée jusqu'à la fin ? Plusieurs stratégies (Schwerpunkt, Guérilla) reposent entièrement sur ce choix.

## A2 : Implémenter les conditions de victoire
Ca devrait être plutôt rapide : 
Maintenant que nous avons nos armées, ils faut décider ce qui met fin à la simulation. J'ai deux propositions : 
- Elimination des unités ennemies sous un certain seuil (je crois que statistiquement, une armée perd une bataille quand plus de 60% de ses effectifs sont éliminés (déroute), mais c'est pour les grandes armées, on pourra régler ce paramètre à la hausse.)
- Capturer et garder un point stratégique pendant un certain temps.

Si vous avez d'autres propositions, n'hésitez pas à les rajouter.

> A fixer :
- **Point stratégique** : comment est-il placé (au centre de la map, à égale distance des spawns pour l'équité) ? Qu'est-ce qui compte comme "capturer" (au moins une unité dessus et aucun ennemi à proximité ?) ? Combien de tours faut-il le tenir ?
- **Fin de partie** : il faut une limite de tours avec match nul, sinon une partie peut ne jamais se terminer (par exemple deux armées en Guérilla qui s'évitent indéfiniment).

## A3 : Implémenter les sous-types d'unités

Nous avons normalement à ce stade les classes filles `soldier` et `support`. Je pense que nous devons maintenant créer les sous-types de ces deux classes : 
### Soldier
Comme vous le verrez, j'ai rajouté la notion de coût : Chaque armée devra avoir le même budget, elle dépensera donc son budget à la création pour définir quels sous-types elle possèdera (3 tank ou 6 soldats par exemple)
- Sous-type -> `infantry` : 
    - Soldat de base
    - PV faible
    - vision moyenne
    - dégâts moyen
    - portée de tir moyenne
    - capacité de munitions moyenne
    - coût moyen
    - transportable
    - déplacement faible
- Sous-type -> `Sniper` : 
    - Soldat à longue distance
    - PV faible
    - vision élevée
    - dégâts élevés 
    - portée de tir longue
    - capacité de munitions faible
    - coût moyen
    - transportable
    - déplacement faible
- Sous-type -> `Tank` : 
    - Soldat blindé
    - PV élevés
    - vision moyenne
    - dégâts élevés
    - portée de tir moyenne
    - capacité de munitions faible
    - coût élevé
    - non transportable
    - déplacement moyen

Comme on a pas encore de référence numérique, j'utilise les mots faible/moyen/élevés pour qu'on ait la référence comparative en tête.

> A fixer :
- **Valeurs chiffrées** : il faudra un tableau de valeurs pour chaque sous-type (PV, dégâts, portée, munitions, coût, déplacement) et le montant du budget de chaque armée.

### Support
- Sous-type -> `Supply Truck` : 
    - PV moyen
    - Déplacement moyen
    - Recharge les munitions des unités à proximité
    - Coût élevé
> A fixer : les supply trucks peuvent recharger les munitions des soldats à l'infini ? Ou faut-il décider que les armées ont une base et que les Supply trucks doivent des allers-retours à la base pour se ravitailler (dans ce cas on touche au concept de ligne de ravitaillement)

- Sous-type -> `Transport Truck`
    - PV moyen
    - Déplacement moyen
    - Capacité de transport élevée
    - Coût moyen
    - Les soldats transportables peuvent rentrer à l'intérieur et se faire déplacer
> A fixer (valable aussi pour l'hélico) : comment monte-t-on et descend-on (depuis une case adjacente ?) ? Que deviennent les passagers si le véhicule est détruit ? Peuvent-ils tirer depuis l'intérieur ?

- Sous-type -> `Transport Helicopter`
    - PV faible
    - Déplacement élevé, ignore les obstacles
    - Capacité de transport faible
    - Coût élevé
    - Les soldats transportables peuvent rentrer à l'intérieur et se faire déplacer
    - Attaquable seulement par les unités longue portée
> Remarque : pour le code, il suffit de faire en sorte que pour l'hélico, toutes les cases sont free tant qu'il vole (et tant pis on compte pas les collisions d'hélicos). Si on veut pousser le concept, il faut ajouter une map aérienne au moteur.


## A4 : Implémenter les stratégies militaires

Maintenant que nous avons toute la base complète du jeu, on peut passer à l'implémentation des comportements macro des armées, et du comportement micro qu'elles attribueraient à leurs unités.

> A noter : on enlèvera sûrement la mécanique de spawn aléatoire, donc la mécanique de regroupement au début, car ça rajoute trop de paramètres aléatoire dans les batailles. Les spawns des unités se feront autour du général au début de la partie, avec SpawnAlliedAt.

### Prérequis communs aux stratégies
Les stratégies ci-dessous supposent des mécanismes qui n'existent pas encore dans le code :
- **Attribution de rôles** : un système pour que l'armée donne un rôle micro à chaque unité (Scoot, appât, marteau, enclume...) et puisse le changer en cours de partie.
- **Sous-groupes** : pour le marteau et l'enclume, les cellules de Guérilla, le détachement de blocage de OODA... C'est la stratégie `Squad`, déclarée dans le code mais pas encore implémentée.
- **Mémoire des ennemis** : aujourd'hui, l'armée ne retient que le terrain (`Army.Known`) et une seule cible (`Army.Enemy`). OODA et Schwerpunkt ont besoin de retenir tous les ennemis vus, avec leur type, leur dernière position connue et le tour où ils ont été vus.
- **Formations** : "les tanks devant, l'infanterie sur les côtés" demande une logique de maintien de formation. Aujourd'hui, chaque unité calcule son chemin seule vers l'objectif de l'armée.

### Stratégies individualistes de référence
Pour comparer stratégies individualistes et coordonnées (l'objectif de départ du projet), il faut aussi décrire des stratégies où chaque unité décide seule, sans armée : au minimum une stratégie "chacun pour soi" (chaque unité explore et attaque le premier ennemi qu'elle voit), qui servira de point de comparaison pour toutes les stratégies macro.

### Macro : Attaque simpliste
Après un regroupement, l'armée se déplace en mettant ses combattants en premier et attaque de manière conjointe (en concentrant les tirs) sur les ennemis qu'ils croisent. En ce qui concerne la cible, un ordre de priorité sera attribué en fonction du danger ou de la rapidité d'élimination, mais en respectant cette règle pour l'instant : General>Soldier>Support.
#### Micro : Tank 
Les tanks ayant le plus de PV, ils seront mis à l'avant de l'armée
#### Micro : infanterie
L'infanterie sera sur les côtés et derrière l'armée, en cas d'attaque ou de détection d'ennemis sur les flancs ou à l'arrière.
#### Micro : snipers
Ce type d'armée ne devrait pas avoir beaucoup de snipers, mais en avoir quand même pour affronter les hélicos, ils seront à côté du général au centre de l'armée, avec les camions de ravitaillement.

Cette armée n'utilisera pas de moyens de transports, vu qu'elle bouge de manière groupée.

### Macro : Blitzkrieg
Après regroupement, toutes les soldats rentrent dans des camions de transports et avancent vers un point où l'ennemi a été aperçu, avec une avant-garde de tank. Une fois arrivé au point, les soldats et les tanks se déploie dans une zone de contrôle, avec le général et les camions de ravitaillement au centre. Une fois la zone sécurisée (plus d'ennemis en vue). L'armée se regroupe, les soldats reviennent dans les camions, et l'armée se déplace au point d'intérêt suivant.

#### Micro : à voir, je ne sais pas vraiment comment l'armée va organiser ses unités pour l'instant

### Macro : trap

L'armée se rassemble, puis se déplace à un endroit sécurisé (peu de chemins vers le général). Elle se place ensuite en position d'encerclement (arc de cercle)
Des soldats "appats" explore la map, puis une fois qu'un ennemi est en vue, il le fuit et l'attire le point de bataille (point de tenaille ou son armée enveloppe l'ennemi). Une fois l'ennemi engagé, l'armée entière attaque l'ennemi en se rabattant dessus.

oooooo
o                            <--
o               <--o        <--<Armée ennemie>
o                            <--
oooooo

#### Micro : Appat
Soldat d'infanterie normal, ou camion de transport (pour la vitesse)
Explore la map afin de trouver des ennemis, puis rejoint son point de tenaille, en restant dans la vision de l'ennemi.

#### Micro : sniper
Une fois l'endroit sécurisé définit, ils se placent en position arrière de l'armée où ils peuvent atteindre la majorité des cases "target" (encerclées par l'armée).

#### Micro : Scoot
Soldat d'infanterie normal, ou camion de transport (pour la vitesse).
Unité chargée d'explorer la map afin d'enrichir la carte partagée de l'armée. Priorise les chemins non-explorés dans sa stratégie de déplacement. Une fois l'endroit sécurisé identifié, redevient un soldat normal, ou se fait assigner un autre micro rôle.


> A partir d'ici, je n'invente plus les stratégies, elles viennent de stratégies historiques documentées.
### Macro : Marteau et enclume

> Inspiration : Alexandre le Grand à Gaugamèles (331 av. J.-C.). La phalange fixe l'armée perse pendant que la cavalerie la contourne et la frappe sur le côté.

L'armée se divise en deux groupes. L'**enclume**, lente et solide, se place face à l'ennemi et l'engage pour le fixer sur place. Le **marteau**, mobile, fait un large détour hors de la vision ennemie (grâce à la carte partagée) pour venir le frapper à l'arrière ou sur le flanc. L'ennemi se retrouve pris entre les deux.

Déroulé :
1. Un Scoot trouve l'ennemi.
2. L'enclume avance et se positionne face à lui, à portée de tir maximale, puis tire sans avancer : le but est que l'ennemi reste engagé contre elle et ne voie pas venir le marteau.
3. Le marteau contourne en évitant les cases vues par l'ennemi, et signale à l'armée quand il est en position.
4. Le marteau attaque par l'arrière, les deux groupes concentrent leurs tirs (priorité General>Support>Soldier).

Si le marteau est repéré trop tôt, ou si l'enclume perd trop d'effectifs avant son arrivée, l'enclume recule lentement en tirant, le marteau la rejoint, et l'armée repasse en Attaque simpliste.

```
            Marteau
              |
              v
<Enclume> --> <Armée ennemie>
```

#### Micro : Enclume - infanterie ou tank je dirais
Ligne de front de l'enclume. Tient sa position et tire, mais n'avance pas : son rôle est de retenir l'ennemi, pas de le détruire.

#### Micro : Enclume - sniper
Derrière l'infanterie de l'enclume. Ciblent en priorité les snipers ennemis et les hélicos (qui pourraient repérer le marteau).

#### Micro : Enclume - général et camions de ravitaillement
Derrière l'enclume, hors de portée ennemie. Les camions de ravitaillement ne servent que l'enclume : le marteau doit être autonome en munitions le temps de la manœuvre.

#### Micro : Marteau - tanks
Force de frappe principale : beaucoup de PV pour encaisser la riposte une fois au contact, beaucoup de dégâts pour faire céder l'ennemi rapidement.

#### Micro : Marteau - camions de transport
Si l'armée n'a pas (ou peu) de tanks, le marteau devient de l'infanterie transportée : les camions contournent puis déposent les soldats à distance de tir de l'arrière ennemi.

#### Micro : Marteau - hélico (variante)
Enveloppement vertical : dépose de l'infanterie directement derrière l'ennemi en ignorant les obstacles. Plus rapide, mais risqué si l'ennemi a des snipers en vue.

### Macro : Boucle OODA

> Inspiration : John Boyd (Observer, S'orienter, Décider, Agir) ; celui qui boucle ce cycle plus vite que l'adversaire le désorganise. Pour la phase d'action : la "position centrale" de Napoléon (campagne d'Italie 1796, campagne de Belgique 1815), qui consiste à se placer entre deux forces ennemies pour les empêcher de se rejoindre et les battre l'une après l'autre.

L'objectif est d'avoir la meilleure connaissance de la map possible, puis d'exploiter cet avantage d'information pour séparer les forces ennemies et les battre une par une. D'après les lois de Lanchester, la puissance d'un groupe croît avec le carré de ses effectifs : battre deux groupes de 5 séparément est bien plus facile que d'affronter un groupe de 10.

Déroulé, répété en boucle tous les X tours :
1. **Observer** : le gros de l'armée reste dans une position sûre pendant que des Scoots rapides partent dans plusieurs directions. On enrichit la carte partagée, et on enregistre chaque ennemi aperçu avec sa position, son type et le tour où il a été vu.
2. **S'orienter** : à partir de ces observations, l'armée identifie les groupes ennemis (unités proches les unes des autres), repère les groupes isolés (loin de tout renfort), et les passages étroits de la map qui séparent ces groupes.
3. **Décider** : l'armée choisit comme cible un groupe isolé sur lequel elle a la supériorité locale (comparaison des effectifs au carré), et le passage étroit à bloquer pour empêcher ses renforts d'arriver.
4. **Agir** : un détachement de blocage tient le passage, la force de frappe attaque le groupe isolé le plus vite possible, puis l'armée retourne à l'étape 1 (les positions ont changé).

Pendant l'action, les Scoots continuent d'observer : l'armée doit voir arriver les renforts ennemis avant que l'ennemi ne réagisse. C'est ça, être "plus rapide dans la boucle".

> A noter : avec seulement deux armées en jeu, "séparer les armées" s'applique aux sous-groupes de l'armée ennemie (Scoots, appâts, marteau, unités retardataires...).

#### Micro : Scoot
Comme le Scoot de la stratégie trap, mais n'arrête jamais d'explorer : une fois la map connue, il surveille les zones vues il y a le plus longtemps et suit les déplacements ennemis. Camion de transport de préférence, pour la vitesse. Si un scoot se fait attaquer (si c'est un camion de transport) les soldats qui sont dedans sortent et protègent le camion.

#### Micro : hélico
Observateur longue distance : rapide et ignore les obstacles, idéal pour couvrir la map. Reste hors de portée de tous les snipers ennemis connus.

#### Micro : Détachement de blocage
Infanterie et snipers. Tient le passage étroit choisi et ne poursuit pas l'ennemi : sa seule mission est de retarder les renforts le temps que la force de frappe finisse sa cible.

#### Micro : Force de frappe
Tanks, et infanterie transportée par camion. Doit éliminer le groupe isolé vite, en concentrant les tirs, avant que les renforts n'arrivent.

#### Micro : général et camions de ravitaillement
Restent dans la position centrale sûre, à égale distance des différentes zones d'action, pour que l'armée puisse basculer d'une cible à l'autre plus vite que l'ennemi ne peut se regrouper.

### Macro : Schwerpunkt

> Inspiration : doctrine allemande du *Schwerpunkt* (concentrer tout l'effort sur le point décisif), appliquée au "centre de gravité" de Clausewitz : la source de puissance de l'ennemi dont la perte fait s'effondrer tout le reste.

Dans notre simulation, le centre de gravité d'une armée ennemie est son **général** (s'il meurt, l'armée perd sa coordination) ou ses **camions de ravitaillement** (sans munitions, elle ne peut plus combattre). L'armée explore la map pour les localiser le plus vite possible, puis y concentre toute sa force en ignorant le reste.

Déroulé :
1. **Recherche** : les Scoots ne cherchent pas à couvrir toute la map, mais à trouver le cœur de l'armée ennemie. Dans la plupart de nos stratégies, le général et les camions de ravitaillement sont placés au centre de l'armée : dès qu'un groupe ennemi est repéré, son centre devient la zone prioritaire à observer.
2. **Percée** : une fois la cible localisée (priorité général > camion de ravitaillement), toute l'armée s'y dirige. Elle n'engage pas les autres ennemis sauf s'ils bloquent le chemin, et choisit un trajet qui évite au maximum les zones ennemies connues.
3. **Frappe** : tir concentré exclusivement sur la cible.
4. **Exploitation** : si le général est mort, l'armée ennemie est désorganisée et on repasse en Attaque simpliste pour finir le travail. Si ce sont les camions de ravitaillement, on évite le contact et on attend que l'ennemi soit à court de munitions.

Si l'armée subit trop de pertes avant d'atteindre sa cible (seuil à définir), elle abandonne et se replie : la percée a échoué, et continuer ne ferait qu'empirer la situation.

#### Micro : Scoot
Se dirige vers les zones où des ennemis ont été vus, observe à distance maximale sans jamais engager, et signale tout général ou camion de ravitaillement aperçu.

#### Micro : snipers
Unité clé de cette stratégie : longue portée et gros dégâts, ils peuvent abattre le général depuis l'extérieur de la portée de l'infanterie ennemie. Cherchent une position avec ligne de vue sur la cible et ne tirent que sur elle.

#### Micro : tanks
Fer de lance de la percée : ouvrent un chemin à travers les unités ennemies qui protègent la cible, et encaissent les tirs à la place des snipers.

#### Micro : infanterie
Protège les flancs du couloir ouvert par les tanks, et escorte les snipers.

#### Micro : hélico
Commando : dépose un ou deux snipers ou fantassins directement à côté de la cible. Les pertes sont acceptables si la cible tombe.

#### Micro : général et camions de ravitaillement
Suivent à distance derrière la percée, hors de portée.

### Macro : Guérilla

> Inspiration : la formule en 16 caractères de Mao Zedong (« l'ennemi avance, nous reculons ; l'ennemi s'arrête, nous le harcelons ; l'ennemi s'épuise, nous attaquons ; l'ennemi recule, nous le poursuivons ») et T. E. Lawrence pendant la révolte arabe (1916-1918), qui harcelait le chemin de fer du Hedjaz plutôt que d'affronter l'armée ottomane de face.

Armée légère qui refuse la bataille rangée : elle gagne par usure, en infligeant des pertes sans en subir. C'est une stratégie adaptée à la victoire par déroute (seuil de pertes), mais faible pour tenir un point stratégique.

L'armée se divise en petites cellules indépendantes (2-3 unités : un sniper et un ou deux fantassins), réparties sur la map et partageant la même carte. L'armée attribue une zone à chaque cellule, qui décide ensuite seule de ses actions selon ces 4 règles :
- **L'ennemi avance → nous reculons** : si un groupe ennemi supérieur en nombre approche, la cellule se retire hors de portée, de préférence derrière des obstacles pour casser la ligne de vue.
- **L'ennemi s'arrête → nous le harcelons** : si un ennemi reste immobile, les snipers tirent à portée maximale puis changent de position aussitôt, pour que la "dernière position connue" retenue par l'ennemi soit fausse (ça exploite directement notre logique de poursuite).
- **L'ennemi s'épuise → nous attaquons** : si un groupe ennemi est affaibli (PV ou munitions bas) ou isolé, les cellules proches se regroupent et l'attaquent en concentrant les tirs.
- **L'ennemi recule → nous le poursuivons** : on suit le repli ennemi en tirant sur les unités retardataires.

Priorité des cibles : camions de ravitaillement > unités isolées > général (si l'occasion se présente).

#### Micro : sniper
Cœur de la cellule : tire puis se déplace, et ne reste jamais deux tours au même endroit après avoir tiré si un ennemi est en vue.

#### Micro : infanterie
Escorte le sniper, tend des embuscades sur les passages étroits empruntés par l'ennemi, et décroche la première quand l'ennemi avance.

#### Micro : camion de transport
Extraction rapide des cellules en danger, et repositionnement d'une zone à l'autre.

#### Micro : hélico (optionnel)
Dépose une cellule derrière les lignes ennemies, près de ses camions de ravitaillement.

#### Micro : tanks
Aucun : trop chers, trop lents et impossibles à cacher. Le budget sert à acheter un maximum d'unités légères.

#### Micro : général
Reste caché loin des combats et change souvent de position, puisque c'est la cible naturelle d'une armée ennemie en Schwerpunkt.

## A5 : Protocole d'expérimentation

L'objectif du projet est de comparer des stratégies (individualistes contre coordonnées, et les stratégies macro entre elles). Pour que les comparaisons aient un sens, il faut pouvoir lancer beaucoup de parties dans des conditions contrôlées :
- **Mode sans affichage** : lancer des centaines de parties à la suite sans `PrintWorld` ni le `Sleep` d'une seconde entre les tours.
- **Graine aléatoire configurable** : aujourd'hui, `utils.rng` est initialisé avec l'heure, donc aucune partie n'est reproductible. Il faut pouvoir fixer la graine pour rejouer exactement une partie (utile aussi pour le débogage).
- **Cartes équitables** : soit des cartes symétriques, soit chaque carte jouée deux fois en inversant les camps, sinon le hasard des obstacles et des positions de départ fausse les résultats. Et garantir la connexité (sinon deux armées séparées donnent un nul)
- **Métriques à enregistrer** : vainqueur et condition de victoire, nombre de tours, pertes de chaque camp (par sous-type), munitions et budget restants...
- **Matrice des affrontements** : chaque stratégie contre chaque autre, sur N parties, pour obtenir des taux de victoire comparables. Avec le budget identique, on s'attend à un équilibre de type pierre-feuille-ciseaux entre les compositions d'armée, qu'il sera intéressant de vérifier.
- **Export des résultats** (CSV par exemple) pour pouvoir faire des graphiques.
- **Modes d'exécution** (un binaire, flags) : `--headless` (parties en parallèle entre elles), `--gui`, `--seed N` dans les deux modes.
- **Mode graphique (ebiten)** : pause, pas à pas, vitesse, réinitialisation ; unités colorées par armée avec forme par sous-type, barres de PV et de munitions, overlays (vision, carte connue de l'armée, cible concentrée, objectif), clic sur une unité pour voir son rôle et son action.

## Pour plus tard

- **Hauteur** : ajouter une carte des hauteurs, qui améliore la vision des unités en hauteur et bloque davantage les lignes de vue. Ça rendrait possibles des stratégies comme la contre-pente.
- **Communication** : aujourd'hui, la carte partagée de l'armée est mise à jour instantanément et parfaitement. On pourrait ajouter une portée ou un délai de communication avec le général.
- **Organisation** : définir l'ordre de priorité entre les tâches, qui fait quoi, et ce qui constitue une version minimale livrable (par exemple : A1 + A2 + une stratégie individualiste + Attaque simpliste + A5).

