# UNMAINTAINED

Du to the poor quality of the english corpus data structure, I have lost interest in the project and I'm leaving it here as is as a reference for others developers. Feel free to reuse the code or part of the parser.

Bye and thanks for all the fish.


# wikiquote-parser

Wikiquote XML Parser.

Preliminary version for the frwikiquote multistream xml. Written in go-lang.

# Letter of intention

- Have the ability to extract the quote from the wikiquote XML dump and structure that for further import.
- Isolate and create stats about the usage of the various markup inside the wikiquote project

# Markup Manipulation

## Source

Sample quote and surroundings
```
<page>
  <title>Amour</title>
  <ns>0</ns>
  <id>1865</id>
  <revision>
    ...
    <text xml:space="preserve">{{en cours.CM}}
{{citation|citation=Aimer, c’est trouver sa richesse hors de soi.}}
{{Réf Livre|titre=Éléments de philosophie (Gallimard)|auteur=Alain |éditeur=Librairie Larousse (Dictionnaire des citations françaises et étrangères)|année=1980 |page=3|isbn=2-03-340809-4}}
{{Choisie citation du jour|puce=*|année=2010|mois=mars|jour=1|commentaire=|}}
```

Note that there is no formal linking between the quote and its reference, except from the spatial proximity. This already makes me sad.

## Intended Output

**Quote Object**:
- text: `Aimer, c’est trouver sa richesse hors de soi.`
- author: `Alain`
- booktitle: `Éléments de philosophie (Gallimard)`
- bookeditor: `Librairie Larousse (Dictionnaire des citations françaises et étrangères)`
- bookyear: `1980
- bookpage: `3`
- isbn: `2-03-340809-4`
- topic: `Amour`

# Usage

Read the source and adapt.

- domenech selects quotes based on their structure (look at multiplex())
- hodgson does almost the same

```
go run domenech/main.go
```

# License

MIT, cf License file.
