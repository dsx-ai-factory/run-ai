/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

export { isExtendedResourceName, isValidExtendedResourceQuantity, isValidK8sKey };

const HUGEPAGES_PREFIX = "hugepages-";

function isExtendedResourceName(name: string): boolean {
  if (name.startsWith(HUGEPAGES_PREFIX)) {
    return _isHugepagesSizeValid(name);
  }

  if (_isNativeResource(name) || name.startsWith("requests.")) {
    return false;
  }

  const nameForQuota = "requests." + name;
  const isQualified = _isQualifiedName(nameForQuota);
  return isQualified;
}

function _isNativeResource(name: string): boolean {
  return !name.includes("/") || _isPrefixedNativeResource(name);
}

function _isPrefixedNativeResource(name: string): boolean {
  return name.includes("kubernetes.io/");
}

function _isQualifiedName(value: string): boolean {
  const maxLength = 253;
  const nameRegex = /^([A-Za-z0-9][-A-Za-z0-9_.]*)?[A-Za-z0-9]$/;
  const parts = value.split("/");
  if (parts.length !== 2) return false;
  const prefix = parts[0];
  const name = parts[1];
  if (!prefix.length || !name.length || !nameRegex.test(name) || name.length > maxLength) return false;
  const isDomainValid = _isSubdomainValid(prefix);
  return isDomainValid;
}

function _isSubdomainValid(value: string): boolean {
  const maxLength = 253;
  const domainRegex = /^[a-z0-9]([a-z0-9\-]{0,61}[a-z0-9])?(\.[a-z0-9]([a-z0-9\-]{0,61}[a-z0-9])?)*$/; //eslint-disable-line
  if (value.length > maxLength || !domainRegex.test(value)) return false;
  return true;
}

function isValidExtendedResourceQuantity(value: string): boolean {
  const quantityRegex = /^(\d+(\.\d+)?([nmuμkKMGTPEZY]|m)(i?)?|\d+(\.\d+)?)$/;
  return quantityRegex.test(value);
}

function isValidK8sKey(key: string): boolean {
  const dnsSubdomain = /^(?!-)[a-z0-9]([-a-z0-9]*[a-z0-9])?(\.[a-z0-9]([-a-z0-9]*[a-z0-9])?)*$/;
  const labelKey = /^[a-z0-9]([-a-z0-9_.]*[a-z0-9])?$/;

  if (key.length > 253) return false;

  const parts = key.split("/");
  if (parts.length > 2) return false; // More than one `/` is not allowed.

  if (parts.length === 2) {
    const [prefix, name] = parts;
    return dnsSubdomain.test(prefix) && labelKey.test(name);
  }

  return labelKey.test(parts[0]);
}

function _isHugepagesSizeValid(name: string): boolean {
  const suffix = name.substring(HUGEPAGES_PREFIX.length);
  const hugepagesSizeRegex = /^[0-9]+(Ki|Mi|Gi|Ti|Pi|Ei)$/;
  return hugepagesSizeRegex.test(suffix);
}
