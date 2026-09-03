<#macro emailLayout>
<html lang="${locale.language}">
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
            width: auto;
            max-width: 100%;
            height: 24px;
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
            background: linear-gradient(0deg, rgba(255, 255, 255, 0.30) 0%, rgba(255, 255, 255, 0.30) 100%), #65A800;
        }
        .highlight {
            font-weight: 700;
        }
        @media (max-width: 600px) {
            .card {
                padding: 20px;
            }
            .logo {
                height: 21px;
            }
        }
    </style>
</head>
<body text="#000000" bgcolor="#f2f3f4">
    <div class="container">
        <div class="card">
            <div class="header">
                <a href="https://www.nvidia.com/en-eu/software/run-ai">
                    <img src="https://s3.amazonaws.com/www.run.ai/OS_NVIDIA_Runai_black.svg" class="logo" alt="NVIDIA Run:ai" />
                </a>
            </div>
            <div class="content">
                <#nested>
            </div>
        </div>
    </div>
</body>
</html>
</#macro>
