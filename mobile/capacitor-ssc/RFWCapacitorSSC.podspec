Pod::Spec.new do |s|
  s.name = 'RFWCapacitorSSC'
  s.version = '0.0.0-dev'
  s.summary = 'Native iOS SSC transport for RFW Capacitor applications.'
  s.license = { :type => 'AGPL-3.0-only' }
  s.homepage = 'https://github.com/rfwlab/rfw'
  s.author = 'RFW contributors'
  s.source = { :git => 'https://github.com/rfwlab/rfw.git', :tag => s.version.to_s }
  s.source_files = 'ios/Sources/RFWSSCPlugin/**/*.{swift,h,m,c,cc,mm,cpp}'
  s.ios.deployment_target = '15.0'
  s.dependency 'Capacitor', '>= 8.0', '< 9.0'
  s.swift_version = '5.9'
end
