benchmark:
	bash scripts/benchmarks/run_all.sh

figures:
	python3 scripts/analysis/analyze.py

artifact:
	bash scripts/artifact/reproduce.sh

clean:
	rm -rf benchmarks/results/*
	rm -rf benchmarks/figures/*
