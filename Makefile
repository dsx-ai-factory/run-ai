build:
	@echo "NA"

test:
	@echo "NA"

lint:
	@echo "NA"

validate:
	@echo "Validating CODEOWNERS..."
	@python3 .github/scripts/validate-codeowners.py CODEOWNERS
	@echo "CODEOWNERS validation completed successfully"
