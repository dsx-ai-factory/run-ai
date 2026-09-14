{{- /*
SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
SPDX-License-Identifier: Apache-2.0
*/ -}}
MIME-Version: 1.0
Content-Type: text/html; charset=UTF-8
Content-Transfer-Encoding: quoted-printable

<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="utf-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <style>
        .link {
            text-decoration:none;
            color: #0654fe;
        }
        .link:hover {
            text-decoration: underline;
        }
        body {
            margin: 0;
            padding: 0;
            font-family: 'Roboto', Arial, sans-serif;
        }
        .container {
            text-align: center;
            color: #c1c2c3;
            background-color: #f2f3f4;
            padding: 50px 20px;
            min-height: 100vh;
        }
        .card {
            background-color: #ffffff;
            max-width: 600px;
            width: 100%;
            display: inline-block;
            padding: 30px;
            text-align: left;
            font-size: 14px;
            font-weight: 400;
            color: #5f5f5f;
            line-height: 1.75;
            font-family: 'Roboto', Arial, sans-serif;
            padding-top: 20px;
            box-sizing: border-box;
        }
        .header {
            border-bottom: 1px solid #efefef;
            padding-bottom: 10px;
            margin-bottom: 20px;
        }
        .logo {
            width: 250px;
            max-width: 100%;
            height: auto;
        }
        .content {
            padding-top: 20px;
            padding-bottom: 20px;
        }
        .button-area {
            text-align: left;
            margin: 10px 0;
        }
        .cta-button {
            background-color: #65A800;
            display: inline-flex;
            padding: 10px 16px;
            color: white !important;
            border-radius: 3px;
            font-size: 16px;
            border: none;
            text-decoration: none;
            font-weight: 400;
            justify-content: center;
            align-items: center;
            gap: 10px;
            line-height: 1.2;
            box-sizing: border-box;
            box-shadow: 0 2px 4px 0 rgba(0, 0, 0, 0.20);
        }
        .cta-button:hover {
            background-color: #93C24D;
        }
        .highlight {
            font-weight: 700;
        }
        @media (max-width: 600px) {
            .card {
                padding: 20px;
            }
            .logo {
                width: 200px;
            }
        }
    </style>
</head>
<body text="#000000" bgcolor="#f2f3f4">
    <div class="container">
        <div class="card">
            <div class="header">
                {{ if .MetaProperties.useAttachments }}
                    <a href="https://www.nvidia.com/en-eu/software/run-ai"><img src="cid:run-ai.png" height="24" alt="NVIDIA Run:ai"></a>
                {{ else }}
                    <a href="https://www.nvidia.com/en-eu/software/run-ai"><img src="https://s3.amazonaws.com/www.run.ai/OS_NVIDIA_Runai_black.svg" height="24" alt="NVIDIA Run:ai"></a>
                {{ end }}
            </div>
            <div class="content">
    Hi {{.EventProperties.user_id }},
    <br /><br />
    You've been invited to start using NVIDIA Run:ai.
    <br />
    Use the button below to sign in and set a new password.
    <br /><br />
    <div class="button-area">
        <a href="{{.EventProperties.link }}" class="cta-button" style="color: white; text-decoration: none;">SIGN IN TO NVIDIA RUN:AI</a>
    </div>
    <br />
     <div class="highlight">Your invitation is valid for the next 7 days.</div>
    For more information, check out our <a class="link" href="https://run-ai-docs.nvidia.com/saas">documentation</a> or <a class="link" href="https://www.nvidia.com/en-eu/support/enterprise/#contact-us">contact support</a>
    <br/><br />
    Thanks,
    <br />
        The <a class="link" href="https://www.nvidia.com/en-eu/software/run-ai">NVIDIA Run:ai</a> team
            </div>
        </div>
    </div>
</body>
</html>

{{ if .MetaProperties.useAttachments }}
--boundary-email-0e3f31f02b3ff9267cd
Content-Type: image/png
Content-Transfer-Encoding: base64
Content-Disposition: attachment; filename=run-ai.png
Content-ID: <run-ai.png>
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
--boundary-email-0e3f31f02b3ff9267cd
{{ end }}
