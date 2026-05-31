#!/usr/bin/env bash

set -e

cd papers/latex

pdflatex main.tex
bibtex main
pdflatex main.tex
pdflatex main.tex
