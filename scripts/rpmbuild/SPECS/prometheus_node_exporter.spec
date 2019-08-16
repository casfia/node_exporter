Name:       prometheus-node-exporter
Version:    %VERSION%
Release:    0
Group:      Application
Vendor:     1x
License:    Apache-2
Summary:    prometheus-node-exporter rpm install.
Source0:    node_exporter
Source1:    node_exporter.service
Source2:    node_exporter.logrotate
Source3:    license
Source4:    notice

%description
rpm install node_exporter,

%prep


%install
rm -rf %{buildroot}

mkdir -p %{buildroot}/usr/bin
mkdir -p %{buildroot}/etc/prometheus
mkdir -p %{buildroot}/etc/init.d/
mkdir -p %{buildroot}/etc/logrotate.d/
mkdir -p %{buildroot}/usr/share/doc/prometheus/

chmod u+x %{SOURCE0}
chmod u+x %{SOURCE1}
chmod u+x %{SOURCE2}
cp -Rf %{SOURCE0} %{buildroot}/usr/bin/
cp -Rf %{SOURCE1} %{buildroot}/etc/init.d/node_exporter
cp -Rf %{SOURCE2} %{buildroot}/etc/logrotate.d/node_exporter
cp -Rf %{SOURCE3} %{buildroot}/usr/share/doc/prometheus/license
cp -Rf %{SOURCE4} %{buildroot}/usr/share/doc/prometheus/notice


%files
/usr/bin/node_exporter
/etc/init.d/node_exporter
/etc/logrotate.d/node_exporter
/usr/share/doc/prometheus/license
/usr/share/doc/prometheus/notice

%clean
rm -rf $RPM_BUILD_ROOT