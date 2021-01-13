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
source5:    prometheus_pusher.py
source6:    prometheus_pusher.service
source7:    config.ini

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
chmod u+x %{SOURCE5}
chmod u+x %{SOURCE6}
cp -Rf %{SOURCE0} %{buildroot}/usr/bin/
cp -Rf %{SOURCE1} %{buildroot}/etc/init.d/node_exporter
cp -Rf %{SOURCE2} %{buildroot}/etc/logrotate.d/node_exporter
cp -Rf %{SOURCE3} %{buildroot}/usr/share/doc/prometheus/license
cp -Rf %{SOURCE4} %{buildroot}/usr/share/doc/prometheus/notice
cp -Rf %{SOURCE5} %{buildroot}/usr/bin/
cp -Rf %{SOURCE6} %{buildroot}/etc/init.d/prometheus_pusher
cp -Rf %{SOURCE7} %{buildroot}/etc/prometheus


%files
/usr/bin/node_exporter
/etc/init.d/node_exporter
/etc/logrotate.d/node_exporter
/usr/share/doc/prometheus/license
/usr/share/doc/prometheus/notice
/usr/bin/prometheus_pusher.py
/etc/init.d/prometheus_pusher
/etc/prometheus/config.ini

%clean
rm -rf $RPM_BUILD_ROOT
