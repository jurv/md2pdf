# Spécifications détaillées - Outil `md2pdf`

## 1. Contexte et objectif du document

Ce document complète le cahier des charges initial (`cdc.md`) et le transforme en spécification exploitable pour
l'implémentation. L'objectif est de définir, sans ambiguïté, le comportement attendu de `md2pdf`, les contraintes
techniques, les interfaces publiques et les critères de validation.

L'outil visé est un programme en ligne de commande, destiné à générer des PDF de qualité à partir de Markdown, tout en
remplaçant les scripts hétérogènes utilisés aujourd'hui dans les projets.

## 2. Périmètre produit

La version 1 couvre l'ensemble des besoins fonctionnels identifiés. Il n'y a pas de découpage prévu en "V2". En
conséquence, les fonctions principales (génération mono-source et multi-sources, personnalisation du rendu, gestion des
assets, support PlantUML, cascade de configuration) et les fonctions utilitaires (fusion et compression PDF) sont toutes
incluses dans le périmètre initial.

Les éléments explicitement hors périmètre sont les suivants : interface graphique, installation automatique des
dépendances système, et moteur de rendu alternatif hors chaîne Pandoc.

## 3. Public cible et cas d'usage

`md2pdf` cible les équipes qui rédigent leurs contenus en Markdown et qui doivent produire des documents PDF prêts à
diffuser (documentation utilisateur, spécifications, comptes rendus, suivi projet). L'outil doit fonctionner aussi bien
en poste développeur qu'en environnement CI.

Le flux principal attendu est simple : un utilisateur fournit un document source (ou une collection de sources),
applique une configuration éventuelle, puis génère un PDF final avec une mise en forme cohérente et reproductible.

## 4. Exigences fonctionnelles

### 4.1 Génération PDF mono-source

L'outil doit générer un PDF à partir d'un unique fichier Markdown. Si aucun chemin de sortie n'est fourni, le fichier
PDF est écrit dans le même dossier que la source, avec le même nom de base et l'extension `.pdf`.

La génération doit réussir même en l'absence de front matter, en appliquant des valeurs par défaut neutres.

### 4.2 Génération PDF multi-sources

Le mode multi-sources est piloté par configuration YAML. Deux mécanismes coexistent : une liste explicite ordonnée
(`sources.explicit`) et une sélection automatique par motifs glob (`sources.include`).

L'ordre final de concaténation est strictement défini : les fichiers de `explicit` sont traités en premier dans l'ordre
fourni, puis les fichiers issus de `include` sont ajoutés après tri alphabétique. Les doublons sont supprimés après
normalisation des chemins.

### 4.3 Table des matières et numérotation

La table des matières est configurable avec trois modes : `auto`, `on`, `off`. En mode `auto`, la génération de table
des matières dépend de la structure des titres présents dans le document. Le titre de la table des matières et sa
profondeur maximale doivent être paramétrables.

La numérotation des sections est activable et pilotable via configuration et options CLI, avec priorité au niveau le
plus local.

### 4.4 Métadonnées et habillage documentaire

Le titre, l'auteur, le sujet, ainsi que les éléments d'habillage (logos, en-têtes, pieds de page, couleurs, polices)
doivent être paramétrables. Ces valeurs sont destinées à alimenter à la fois le rendu visuel et les métadonnées du PDF
final.

Les chemins d'assets doivent être robustes : résolution relative au document, puis recherche dans des chemins déclarés
en configuration.

### 4.5 Templates

`md2pdf` fournit un template LaTeX par défaut. L'utilisateur peut le remplacer par un template personnalisé complet via
configuration ou option CLI. Le comportement attendu est binaire : soit le template par défaut est utilisé, soit le
template explicite est pris tel quel.

### 4.6 Support PlantUML

Le support PlantUML est obligatoire lorsqu'un document contient des blocs PlantUML. Si les dépendances nécessaires sont
absentes, la commande doit échouer avec un message explicite indiquant ce qui manque et comment corriger.

En revanche, un document sans bloc PlantUML ne doit pas être bloqué par l'absence de cette dépendance.

### 4.7 Fonctions utilitaires PDF

La CLI doit aussi proposer des commandes dédiées pour fusionner des PDF (`merge`) et compresser un PDF (`compress`). Ces
commandes font partie de la V1 et doivent être documentées au même niveau que la génération Markdown -> PDF.

## 5. Interface en ligne de commande

La CLI est organisée autour de sous-commandes pour conserver une interface lisible et évolutive.

Commandes obligatoires :

- `md2pdf build <input.md> [-o output.pdf] [options]`
- `md2pdf merge <file1.pdf> <file2.pdf> ... -o merged.pdf`
- `md2pdf compress <input.pdf> -o output.pdf [--quality screen|ebook|printer|prepress]`
- `md2pdf doctor [--json]`
- `md2pdf init [--profile default|report|meeting]`

Options minimales de `build` :

- `--config <path>`
- `--project-config <path>`
- `--pdf-engine <xelatex|lualatex|pdflatex>`
- `--template <path>`
- `--toc <auto|on|off>`
- `--toc-title <texte>`
- `--toc-depth <n>`
- `--verbose` / `--debug`

Chaque commande doit afficher une aide claire (`--help`) avec exemples d'usage.

## 6. Modèle de configuration YAML

La configuration repose sur YAML, à la fois pour les fichiers de configuration et pour le front matter. Ce choix
garantit une cohérence d'usage et limite la charge cognitive côté utilisateur.

L'héritage se fait selon trois niveaux : global machine, projet, document. La fusion suit les règles suivantes : fusion
profonde des objets, remplacement des scalaires et des listes, suppression explicite d'une valeur héritée via `null`.

Exemple de structure cible :

```yaml
pdf:
  engine: xelatex
  template: null
metadata:
  title: null
  author: null
  subject: null
toc:
  mode: auto
  title: Sommaire
  depth: 3
sources:
  explicit: []
  include: []
assets:
  search_paths: []
  logo_cover: null
  logo_header: null
style:
  colors:
    primary: "#1F4E79"
  fonts:
    body: "Open Sans"
    heading: "Open Sans"
header_footer:
  header_left: null
  header_right: null
  footer_left: null
  footer_right: null
features:
  plantuml: auto
```

## 7. Architecture technique

L'implémentation cible est en Go, avec le framework Cobra pour la CLI. Cette combinaison répond au besoin de binaire
standalone multi-OS, avec une distribution simple et une maintenance raisonnable.

La chaîne de rendu est basée sur Pandoc, avec moteur PDF configurable (`xelatex`, `lualatex`, `pdflatex`). Le mode par
défaut est `xelatex`, car il est aujourd'hui le plus proche des usages existants.

Les dépendances système ne sont pas installées automatiquement. En revanche, `md2pdf doctor` doit vérifier leur
présence, détecter les versions et fournir des actions correctives explicites.

## 8. Journalisation, erreurs et codes de sortie

Le niveau de logs doit être cohérent sur toutes les commandes (`error`, `warn`, `info`, `debug`) et orienté diagnostic
utilisateur.

Codes de sortie normés :

- `0` : succès
- `2` : erreur d'entrée ou de configuration utilisateur
- `3` : dépendance manquante
- `4` : erreur de rendu ou erreur interne de pipeline

Les messages d'erreur doivent toujours indiquer le contexte (fichier, clé de config, commande ou étape concernée) et,
lorsque possible, une action de correction.

## 9. Exigences de qualité et stratégie de tests

La validation de la V1 repose sur des tests E2E en matrice multi-OS (Windows, Linux, macOS). L'objectif est de garantir
un comportement stable entre environnements et de détecter rapidement les régressions de rendu.

Un jeu de "golden files" PDF doit être maintenu sur des scénarios représentatifs : mono-source simple, mono-source
enrichi, multi-sources, ToC activée/désactivée, template custom, assets manquants, PlantUML présent/absent, fusion PDF,
compression PDF.

La comparaison des sorties doit intégrer une politique de tolérance explicite pour les métadonnées temporelles afin
d'éviter les faux positifs.

## 10. Critères d'acceptation de la V1

La V1 sera considérée comme livrable lorsque toutes les commandes publiques seront opérationnelles, documentées, et
validées par la matrice de tests définie ci-dessus.

En particulier, l'outil devra :

- fonctionner de manière équivalente sur les trois OS cibles ;
- produire un PDF valide sans front matter ;
- appliquer correctement la cascade de configuration YAML ;
- respecter l'ordre de fusion multi-sources ;
- gérer proprement les erreurs de dépendances via `doctor` ;
- exécuter sans ambiguïté les fonctions `build`, `merge` et `compress`.

## 11. Références projet

Les dossiers `examples/` existants constituent la base de vérité fonctionnelle initiale (cas historiques, templates,
scripts et documents types). Les cas de tests de la V1 doivent s'appuyer sur ces exemples pour garantir que le nouvel
outil couvre les besoins réels et supprime la duplication actuelle.
