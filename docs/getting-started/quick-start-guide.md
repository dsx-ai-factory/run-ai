---
hidden: true
---

# NVIDIA Run:ai Quick Start Guide

## Activating the Accounts for Getting NVIDIA Run:ai

After your order for NVIDIA Run:ai is processed, you receive an order confirmation message from NVIDIA. This message contains information to access NVIDIA Run:ai software and get technical support from NVIDIA.

To access these services, you must have an NVIDIA Enterprise Account, which provides login access to the following:

* **NVIDIA NGC** - Provides access to all enterprise software, services, and management tools included in NVIDIA Run:ai.
* **NVIDIA Enterprise Support Portal** - Provides access to NVIDIA Run:ai support services.

You can reach these web properties from the [NVIDIA Application Hub](http://nvid.nvidia.com/dashboard/).

### Prerequisites

Before you begin, ensure that you have a valid NVIDIA software subscription.

### Your Order Confirmation Message

After your order for NVIDIA Run:ai is processed, you receive an order confirmation message to which your NVIDIA Entitlement Certificate is attached. Your NVIDIA Entitlement Certificate contains your product activation keys and provides instructions for using the certificate.

If you are a data center administrator, follow the instructions in the NVIDIA Entitlement Certificate to use the certificate. Otherwise, forward your order confirmation message, including the attached NVIDIA Entitlement Certificate, to a data center administrator in your organization.

## NVIDIA Enterprise Account Requirements

You must have a suitable NVIDIA Enterprise Account for getting NVIDIA Run:ai and technical support from NVIDIA.

Whether you have a suitable NVIDIA Enterprise Account depends on whether you have previously purchased NVIDIA AI Enterprise or other NVIDIA services that require an enterprise account. If you have not previously purchased NVIDIA Run:ai, you do not have a suitable NVIDIA Enterprise Account.

### Creating Your NVIDIA Enterprise Account

Before you begin, ensure that you have your [order confirmation message](#your-order-confirmation-message).

* If you do not have an NVIDIA Enterprise Account, create an account to log in to the web properties for getting NVIDIA Run:ai and technical support from NVIDIA.
* To create a separate account for your purchased licenses, perform this task using a different email address than the address you used to create your existing account.
* If you already have an account, skip this task and go to [Installing NVIDIA Run:ai Software Components](#installing-nvidia-runai-software-components).

To create an NVIDIA Enterprise Account:

1. In the instructions for using your NVIDIA Entitlement Certificate, follow the **Register** link.
2. Fill out the form on the NVIDIA Enterprise Account Registration page and select **REGISTER**.
   * A message confirming that an account was created appears.
   * NVIDIA sends an email to the address you provided, instructing you to log in to your account on the [NVIDIA Application Hub](http://nvid.nvidia.com/dashboard/).
3. Open the email instructing you to log in to your account and select **Log In**.
4. On the NVIDIA Application Hub Login page that opens, in the text-entry field, enter the email address you provided and select **Sign In**.
5. On the Create Your Account page that opens, provide and confirm a password for the account and select **Create Account**.
   * A message prompting you to verify your email address appears.
   * NVIDIA sends an email to the address you provided, instructing you to verify your email address.
6. Open the email instructing you to verify your email address and select **Verify Email Address**. A message confirming that your email address is confirmed appears.

## Installing NVIDIA Run:ai Software Components

### The NVIDIA NGC Catalog

The installation artifacts of NVIDIA Run:ai are available under the NVIDIA Run:ai collections and resources. Before downloading any NVIDIA Run:ai assets, sign in to NVIDIA NGC from the [NVIDIA NGC Sign In](https://ngc.nvidia.com/signin) page.

The NVIDIA Run:ai Support-entitled collection is fully deployed within your own infrastructure, including both the control plane and connected clusters. This deployment model gives you full control over platform configuration, security boundaries, and operational management within your environment. NVIDIA Run:ai self-hosted is included in the NVIDIA AI Enterprise license.

<Note>
**Note**

From version 3.0, there is no ready-made NVIDIA Run:ai download on NGC for air-gapped sites. Instead, the installer builds the package for you on a machine that has internet access.
</Note>

### Installing NVIDIA Run:ai Software

For prerequisites and installation instructions, refer to [Installation and Configuration](../installation-and-configuration/overview.md).

## Notices

### Notice

This document is provided for information purposes only and shall not be regarded as a warranty of a certain functionality, condition, or quality of a product. NVIDIA Corporation ("NVIDIA") makes no representations or warranties, expressed or implied, as to the accuracy or completeness of the information contained in this document and assumes no responsibility for any errors contained herein. NVIDIA shall have no liability for the consequences or use of such information or for any infringement of patents or other rights of third parties that may result from its use. This document is not a commitment to develop, release, or deliver any Material (defined below), code, or functionality.

NVIDIA reserves the right to make corrections, modifications, enhancements, improvements, and any other changes to this document, at any time without notice.

Customers should obtain the latest relevant information before placing orders and should verify that such information is current and complete.

NVIDIA products are sold subject to the NVIDIA standard terms and conditions of sale supplied at the time of order acknowledgement, unless otherwise agreed in an individual sales agreement signed by authorized representatives of NVIDIA and customer ("Terms of Sale"). NVIDIA hereby expressly objects to applying any customer general terms and conditions with regard to the purchase of the NVIDIA product referenced in this document. No contractual obligations are formed either directly or indirectly by this document.

NVIDIA products are not designed, authorized, or warranted to be suitable for use in medical, military, aircraft, space, or life support equipment, nor in applications where failure or malfunction of the NVIDIA product can reasonably be expected to result in personal injury, death, or property or environmental damage. NVIDIA accepts no liability for inclusion or use of NVIDIA products in such equipment or applications and therefore such inclusion or use is at the customer's own risk.

NVIDIA makes no representation or warranty that products based on this document are suitable for any specified use. Testing of all parameters of each product is not necessarily performed by NVIDIA. It is the customer's sole responsibility to evaluate and determine the applicability of any information contained in this document, ensure the product is suitable and fit for the application planned by the customer, and perform the necessary testing for the application to avoid a default of the application or the product. Weaknesses in the customer's product designs may affect the quality and reliability of the NVIDIA product and may result in additional or different conditions or requirements beyond those contained in this document. NVIDIA accepts no liability related to any default, damage, costs, or problem that may be based on or attributable to: (i) the use of the NVIDIA product in any manner that is contrary to this document, or (ii) customer product designs.

No license, either expressed or implied, is granted under any NVIDIA patent right, copyright, or other NVIDIA intellectual property right under this document. Information published by NVIDIA regarding third-party products or services does not constitute a license from NVIDIA to use such products or services or a warranty or endorsement thereof. Use of such information may require a license from a third party under the patents or other intellectual property rights of the third party, or a license from NVIDIA under the patents or other intellectual property rights of NVIDIA.

Reproduction of information in this document is permissible only if approved in advance by NVIDIA in writing, reproduced without alteration and in full compliance with all applicable export laws and regulations, and accompanied by all associated conditions, limitations, and notices.

This document and all NVIDIA design specifications, reference boards, files, drawings, diagnostics, lists, and other documents (together and separately, "Materials") are being provided "as is." NVIDIA makes no warranties, expressed, implied, statutory, or otherwise with respect to the Materials, and expressly disclaims all implied warranties of noninfringement, merchantability, and fitness for a particular purpose. To the extent not prohibited by law, in no event will NVIDIA be liable for any damages, including without limitation any direct, indirect, special, incidental, punitive, or consequential damages, however caused and regardless of the theory of liability, arising out of any use of this document, even if NVIDIA has been advised of the possibility of such damages. Notwithstanding any damages that the customer might incur for any reason whatsoever, NVIDIA's aggregate and cumulative liability toward the customer for the products described herein is limited in accordance with the Terms of Sale for the product.

### VESA DisplayPort

DisplayPort and DisplayPort Compliance Logo, DisplayPort Compliance Logo for Dual-mode Sources, and DisplayPort Compliance Logo for Active Cables are trademarks owned by the Video Electronics Standards Association in the United States and other countries.

### HDMI

HDMI, the HDMI logo, and High-Definition Multimedia Interface are trademarks or registered trademarks of HDMI Licensing LLC.

### OpenCL

OpenCL is a trademark of Apple Inc. used under license to the Khronos Group Inc.

### Trademarks

NVIDIA, the NVIDIA logo, NVIDIA Maxwell, NVIDIA Pascal, NVIDIA Turing, NVIDIA Volta, Quadro, and Tesla are trademarks or registered trademarks of NVIDIA Corporation in the U.S. and other countries. Other company and product names may be trademarks of the respective companies with which they are associated.

### Copyright

Copyright (c) 2024 NVIDIA Corporation and affiliates. All rights reserved.
