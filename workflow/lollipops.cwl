#!/usr/bin/env cwl-runner

cwlVersion: v1.0
class: CommandLineTool
baseCommand: lollipops
hints:
  DockerRequirement:
    dockerPull: pbnjay/lollipops:latest
inputs:
  gene:
    doc: "gene or protein id to query for domain information"
    type: string?
    inputBinding:
      position: 3

  uniprot:
    type: string?
    inputBinding:
      prefix: -U
      position: 2

  querydb:
    doc: "UniprotKB accession database to query"
    name: querydb
    default: GENENAME
    type:
      type: enum
      symbols:
        - GENENAME
        - P_REFSEQ_AC
        - ENSEMBL_ID
        - P_ENTREZGENEID
    inputBinding:
      prefix: -Q
      position: 2

  variants:
    doc: "list of protein variants in amino acid space"
    type: string[]?
    inputBinding:
      position: 4

  synColor:
    doc: "color to use for synonymous mutation markers"
    default: "#0000ff"
    type: string?
    inputBinding:
      prefix: "-syn-color="
      separate: false

  nonsynColor:
    doc: "color to use for non-synonymous mutation markers"
    default: "#ff0000"
    type: string?
    inputBinding:
      prefix: "-mut-color="
      separate: false

  imageDPI:
    doc: "set image DPI (affects PNG output only)"
    type: int
    default: 72
    inputBinding:
      prefix: "-dpi="
      separate: false

  imageWidth:
    doc: "set diagram pixel width (default/0 = automatically size image to fit)"
    type: int
    default: 0
    inputBinding:
      prefix: "-w="
      separate: false

  domainFitMethod:
    doc: "how to apply domain labels"
##   -domain-labels=fit      hot to apply domain labels (default="truncated")
##                             "fit" = only if fits in space available
##                             "off" = do not draw text in the domains
    default: "truncated"
    type:
      type: enum
      symbols:
       - "truncated"
       - "fit"
       - "off"

  hideAxis:
    doc: "do not draw the amino position x-axis"
    type: boolean?
    inputBinding:
      prefix: "-hide-axis"
  showDisordered:
    doc: "draw disordered regions on the backbone"
    type: boolean?
    inputBinding:
      prefix: "-show-disordered"
  showMotifs:
    doc: "draw simple motif regions"
    type: boolean?
    inputBinding:
      prefix: "-show-motifs"
  noPatterns:
    doc: "use solid fill instead of patterns (SVG only)"
    type: boolean?
    inputBinding:
      prefix: "-no-patterns"
  labels:
    doc: "draw label text above lollipop markers"
    type: boolean?
    inputBinding:
      prefix: "-labels"
  legend:
    doc: "draw a legend for colored regions"
    type: boolean?
    inputBinding:
      prefix: "-legend"
  imagename:
    doc: "output image filename (use .png or .svg)"
    type: string
    inputBinding:
      prefix: "-o"

  uniprotDomains:
    doc: "use uniprot domains instead of Pfam"
    type: boolean?
    inputBinding:
      prefix: "-uniprot"
  localDomainFile:
    doc: "get domain info from a file"
    ## see: http://pfam.xfam.org/help#tabview=tab9
    type: File?
    inputBinding:
      prefix: "-l="
      separate: false
    
outputs:
  image:
    type: File
    outputBinding:
      glob: $(inputs.imagename)

s:author:
  - class: s:Person
    s:identifier: https://orcid.org/0000-0002-5761-7533
    s:email: mailto:jeremy@pbnjay.com
    s:name: Jeremy Jay

s:citation: http://dx.doi.org/10.1371/journal.pone.0160519
s:codeRepository: https://github.com/joiningdata/lollipops
s:dateCreated: "2014-09-12"
s:license: https://spdx.org/licenses/GPL-3.0-only

s:keywords: edam:topic_0091 , edam:topic_0622
s:programmingLanguage: Go

$namespaces:
 s: https://schema.org/
 edam: http://edamontology.org/

$schemas:
 - https://schema.org/version/latest/schema.rdf
 - http://edamontology.org/EDAM_1.18.owl
