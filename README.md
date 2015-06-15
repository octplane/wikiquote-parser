# wikiquote-parser

Wikiquote XML Parser.

Preliminary version for the frwikiquote multistream xml. Writtent in go-lang

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

```
go run domenech/main.go
```

Will read sample.xml and output the structured version of what it can find.

# License

MIT, cf License file.
