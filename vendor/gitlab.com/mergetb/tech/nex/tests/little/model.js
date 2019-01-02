topo = {
  name: 'nex-mini',
  nodes: [deb('server'), deb('db'), deb('c0'), deb('c1'), deb('c2')],
  switches: [cumulus('cx')],
  links: [
    Link('server', 1, 'cx', 1),
    Link('db', 1, 'cx', 2),
    Link('c0', 1, 'cx', 3),
    Link('c1', 1, 'cx', 4),
    Link('c2', 1, 'cx', 5),
  ]
}

function deb(name) {
  return {
    name: name,
    image: 'debian-buster',
    cpu: { cores: 4 },
    memory: { capacity: GB(4) },
    mounts: [
      { source: env.PWD+'/../../../nex', point: '/tmp/nex' }
    ]
  }
}

function cumulus(name) {
  return {
    name: name,
    image: 'cumulusvx-3.5-mvrf',
    cpu: { cores: 4 },
    memory: { capacity: GB(4) },
    mounts: [
      { source: env.PWD+'/../../../nex', point: '/tmp/nex' }
    ]
  }
}

