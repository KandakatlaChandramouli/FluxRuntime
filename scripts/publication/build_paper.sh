#!/usr/bin/env bash

cd papers/latex

pdflatex main.tex
bibtex main
pdflatex main.tex
