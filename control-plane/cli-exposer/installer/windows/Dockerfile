FROM mcr.microsoft.com/dotnet/sdk:8.0
WORKDIR /msi

RUN dotnet tool install --global wix

COPY ./installer/windows/media ./media
COPY ../../bin/dist/windows ./bin
COPY ./docs ./docs
COPY ./installer/windows/installer.wxs ./installer.wxs

ENV PATH="${PATH}:/root/.dotnet/tools"
RUN wix extension add WixToolset.Util.wixext
RUN wix extension add WixToolset.UI.wixext

ENTRYPOINT ["wix", "build","-ext" ,"WixToolset.Util.wixext", "-ext", "WixToolset.UI.wixext", "installer.wxs"]