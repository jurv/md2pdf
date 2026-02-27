# md2pdf - Génération de PDF depuis des fichiers Markdown

## Généralité

Afin de faciliter la gestion de projet, la documentation client et interne, et d'autres documents, nous avons adopter le
format markdown. Ce choix s'est fait car :

- c'est un format permettant l'accès au contenu sans aucun logiciel spécial (un simple éditeur suffit)
- il est supporté dans de nombreux outils
- il minimise le stockage
- il est générable et interrogeable facilement par les outils d'IA

En revanche, ce format n'est pas idéal pour la diffusion de l'information à l'extérieur de la société (envoi de
documentation au client, de cahier des charges, de compte-rendu de réunion).

Pour répondre à cette problématique, il est nécessaire d'avoir un outil qui permette la conversion facile de fichier
markdown en fichier PDF

## Historique

Des solutions ont déjà été apportées dans les projets précédent.

### Solution technique

La solution mise en place consiste en une conversion en PDF via Latex.

Pour celà, un script bash est le "chef d'orchestre" de la conversion : il gére la conversion en latex, puis l'export du
PDF depuis le latex intermédiaire.

### Cas initial

Le premier projet (*P1*) concerné était une documentation utilisateur, intégrée dans une application mobile.
La documentation était rédiger en markdown, mais "splitté" en plusieurs fichier pour faciliter la gestion. Le découpage
a été réalisé ainsi :

- `cover.md`
- `000 - login.md`
- `001 - home.md`
- `002 - accound.md`
- ...

La cover était géré à part, afin de pouvoir lui donner une mise en forme particulière (image de fond, taille de police,
pas de header/footer, etc). Elle donnait lieu à son propre PDF. Puis, le reste de la documentation était généré à part,
et la documentation finale consistait en un "assemblage" des deux PDF intermédiaires.

Cela impliquait la gestion de la numérotation des pages (le deuxième PDF commençait en page 2).

Bien que complexe, cette solution présente l'avantage de fournir un PDF unique avec une mise en page de qualité et
esthétiquement plaisante, pouvant être présentée à un utilisateur final de l'application mobile.

### Cas suivant

Dans d'autres projets, nous avons ensuite repris un fonctionnement similaire, mais avons eu besoin de l'adapter.

Dans le cas de documents plus simple, nous n'avons pas de "cover" à mettre en place. Pour des documents très simple, il
n'y a pas besoin de table des matières, etc etc.

Des besoins fonctionnels supplémentaires se sont révélés au fur et à mesure des projets, comme la gestion des schémas
plantuml.

## Problématiques

Bien que fonctionnelle, la solution actuelle présente des problèmes :

- beaucoup d'information sont présentes dans le template latex (couleurs, contenus textuels, lien vers le logo, etc.)
- la gestion des assets est fait via des chemins relatifs dans le latex ou dans le script shell, ce qui pose des
  problèmes (erreur si pas générée depuis le bon dossier, etc)
- duplication : le script, le(s) template(s) LaTeX sont dupliqués entre les projets, et modifié au fur et à mesure des
  projet. Il n'y a donc pas d'unicité, et des fonctionnalités développées à un endroit ne sont pas accessible dans les
  autres projets
- flexibilité : des choses en dur dans les codes et templates, et donc "difficilement" modifiable

## Solution souhaitée

La solution souhaitée serait un outil unique, `md2pdf`, proposant une solution de conversion de markdown en PDF unique.

### Fonctionnelle

#### Fonction principale

Nous envisageons d'utiliser extensivement la notation `front matter` afin de décrire différents éléments
du fichier final souhaité :

- Titre du document : utilisé dans les meta du PDF ?
- Auteur : utilisé dans les meta du PDF ?
- Document multi-source ? : Est-ce que le document est un markdown unique, ou est-ce qu'il est composé de plusieurs
  markdown ?
- Table des matières ? : dans les dernières versions, le besoin de faire apparaître ou non une table des matières était
  déterminer par les titres présent dans le markdown. Il faudrait garder ça comme comportement par défaut, avec
  possibilité de le surcharger ici.
- Titre de la table des matières
- Niveau de titre max de la table des matières
- Liste ou regex des autres sources : si document multi-source, il faudrait trouver un moyen de lister les autres
  documents sources, et l'ordre dans lequel ils doivent être inclus.
- Template tex : il doit y avoir un template tex par défaut, avec possibilité de le "surcharger". Pareil pour la cover
  dans un document multi-source.
- Couleurs : actuellement, une couleur est défini dans le tex pour colorer les titres, les séparateurs, les liens, etc..
- Lien vers logo :
  - pour la page d'accueil
  - pour le header / footer
- Polices : les polices de texte, de titre, de header/footer, etc., ainsi que leur tailles et couleurs pour ces
  différentes parties du fichier.
- ...

La liste est non exhaustive, et d'autres éléments ont peut-être vocation à être ajoutées.

Les contenus des headers / footer pourraient idéalement aussi être configurables.

Pour faciliter la réutilisation, une "cascade" de configuration devrait être mise en place.

Tout ou partie de cette configuration devrait pouvoir être chargé via :

- un fichier de configuration général, pour l'ensemble des projets de la machine
- surchargé par un fichier de configuration partagé dans un projet, 
- surchargé document par document. 

Cela permet de mutualiser les choses qui doivent l'être, permettant d'éviter la duplication et de faciliter l'application globale d'une modification (ex: client
qui change de nom ou de logo).

Tous ces éléments sont optionnels : en l'absence de front matter, la génération doit quand même se faire avec des
valeurs par défaut neutre.

#### Fonctions secondaires

Via des options, la solution obtenue devrait permettre de fournir des outils facilitant la manipulation de fichier PDF ou
markdown :
- compresser des PDF
- fusionner des PDF
- ...

et autres fonctions implémentées dans les solutions manuelles précédentes

### Technique

Dans l'idéal, l'outil devrait se composer d'un programme en ligne de commande, ainsi que d'un fichier de configuration
(emplacement des binaires nécessaire pour le latex ou le pdf, etc...).

Les contraintes techniques sont :
- la portabilité : doit fonctionner sur toutes les plateforme (Windows, Linux, MacOS)
- la simplicité : l'installation et l'utilisation doivent être simplifié au maximum
- la maintenabilité : préféré une structure simple, avec un deboggage facile
- fonctionnement en ligne de commande

Le choix du langage permettant de répondre à ces exigences n'a pas encore été fait.

Il devrait pouvoir prendre en paramètre un fichier.
En paramètre optionnel le fichier de sorti, par défaut dans le même emplacement que le markdown source, avec le même nom
que ce fichier source avec l'extension PDF.

## Annexe

Dans le dossier `example` se trouvent plusieurs exemples de fichier bash, markdown et/ou de fichier latex, ayant utilisé
dans différents projets.

Ces exemples doivent être étudiés afin de valider l'exhaustivité des fonctionnalités listées dans ce cahier des
charges.
