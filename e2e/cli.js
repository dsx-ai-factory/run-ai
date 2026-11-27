#!/usr/bin/env node
/**
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

// used for npm cli-command
// this allows us to use this package as npm command line tool
var exec = require('child_process').exec;
exec('npm run test');
