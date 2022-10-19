FROM mcr.microsoft.com/mssql/server:2019-latest
USER root

#Install dotnet
RUN apt-get update \
   && ACCEPT_EULA=Y apt-get upgrade -y \
   && apt-get install -y wget \
   && wget --no-dns-cache https://packages.microsoft.com/config/ubuntu/18.04/packages-microsoft-prod.deb -O packages-microsoft-prod.deb \
   && dpkg -i packages-microsoft-prod.deb \
   && apt-get update \
   && apt-get install -y dotnet-runtime-3.1 zip \
   && dpkg --purge packages-microsoft-prod \
   && apt-get purge -y wget \
   && apt-get clean \
   && rm packages-microsoft-prod.deb \
   && rm -rf /var/lib/apt/lists/*

#Download and unpack DxE, setup permissions
ADD https://repos.dh2i.com/container/ ./dxe.tgz
RUN tar zxvf dxe.tgz && rm dxe.tgz \
   && chown -R mssql /var/opt/mssql \
   && chmod -R 777 /opt/dh2i /etc/dh2i

#Finish setup
EXPOSE 7979 7985
ENV DX_HAS_MSSQLSERVER=1
USER mssql
ENTRYPOINT ["/opt/dh2i/sbin/dxstart.sh"]