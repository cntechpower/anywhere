Summary: Anywhere By cntechpower
Name: anywhere
Version: 0.0.2
Release: qa
Source0: %{name}.tar.gz
License: Commercial
Group: cntechpower
Prefix: /usr/local/anywhere

%description
Anywhere By cntechpower

##########
%prep
%setup -q

##########
%build
echo "build anywhere..."
cd %{_builddir}/%{buildsubdir}/src
make


##########
%install
rm -rf $RPM_BUILD_ROOT
mkdir -p $RPM_BUILD_ROOT/usr/local/anywhere/bin
ls -l %{_builddir}/%{buildsubdir}/src/bin
cp %{_builddir}/%{buildsubdir}/src/bin/anywhere $RPM_BUILD_ROOT/usr/local/anywhere/bin/anywhere
cp %{_builddir}/%{buildsubdir}/src/bin/anywhered $RPM_BUILD_ROOT/usr/local/anywhere/bin/anywhered

touch $RPM_BUILD_ROOT/usr/local/anywhere/flags


##########
%clean
#rm -rf $RPM_BUILD_ROOT



##########
%post
#chmod
find $RPM_INSTALL_PREFIX -type d -exec chmod 0750 {} \;
find $RPM_INSTALL_PREFIX -type f -exec chmod 0640 {} \;
chmod 0750 $RPM_INSTALL_PREFIX/bin/*

%files
%defattr(-,root,root)
/usr/local/anywhere/bin/anywhered
/usr/local/anywhere/bin/anywhere
%config(noreplace) /usr/local/anywhere/flags