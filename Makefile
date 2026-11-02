# SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
# SPDX-License-Identifier: Apache-2.0

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
