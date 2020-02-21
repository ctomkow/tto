Name:		tto		
Version:	0.5.1	
Release:	1%{?dist}
Summary:	tto package	

Group:		System Environment/Base	
Packager:	Craig Tomkow
License:	MIT	
URL:		https://github.com/ctomkow/tto	
Source0:	tto-0.5.1.tar.gz

Requires:	mariadb

%description
 -- tto
Three backups, two copies on different storage, one located off-site.
An asynchronous client-server app for synchronizing a MySQL database between two systems. In addition, it keeps a ring buffer of X backups on the secondary system.
The main use-case for developing this was to help maintain a hybrid primary / [primary/secondary] application deployment where replication was not possible.

 -- Configuration
The application needs to be installed on the primary and secondary systems. Each will be configured for their respective roles (sender | receiver).
Edit conf.json in /etc/tto/

%prep
%setup -q # unpack tar.gz

%install
mkdir -p %{buildroot}/usr/local/bin/
cp -rfa * %{buildroot}/usr/local/bin/

%files
%attr(0744, root, root) /usr/local/bin/*

%post
/usr/local/bin/tto install

%preun
/usr/local/bin/tto remove

%postun
rm -r /opt/tto/
rm -r /etc/tto/
exit

%doc
%changelog

