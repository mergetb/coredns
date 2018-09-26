
function cumulus(name) {
  return {
    name: name,
    image: 'cumulusvx-3.5-mvrf',
    cpu: { cores: 2 },
    memory: { capacity: MB(512) }
  }
}

function fedora(name, mem) {
  return {
    name: name,
    image: 'fedora-28',
    cpu: { cores: 2 },
    memory: { capacity: MB(mem) },
    mounts: [
      { source: env.PWD+'/../../../nex', point: '/tmp/nex' }
    ]
  }
}

function netboot(name, mem) {
  return {
    name: name,
    image: 'netboot',
    os: 'netboot',
    'no-testnet': true,
    cpu: { cores: 2 },
    memory: { capacity: MB(mem) }
  }
}

topo = {
  name: 'nex0',
  nodes: [
    fedora('s0', 1024),
    fedora('s1', 1024),
    fedora('db', 1024),
    fedora('c0', 512),
    fedora('c1', 512),
    netboot('c2', 512),
    netboot('c3', 512)
  ],
  switches: [
    cumulus('sw')
  ],
  links: [
    Link('s0', 1, 'sw', 1),
    Link('s1', 1, 'sw', 2),
    Link('db', 1, 'sw', 3),
    Link('c0', 1, 'sw', 4),
    Link('c1', 1, 'sw', 5),
    Link('c2', 1, 'sw', 6, {
      boot: 1, 
      mac: {
        c2: '00:00:99:10:00:01'
      }
    }),
    Link('c3', 1, 'sw', 7, {
      boot: 1, 
      mac: {
        c3: '00:00:99:22:00:11'
      }
    }),
  ]
}
