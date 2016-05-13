#!/usr/bin/env python
import os
from distutils.core import setup, Extension
sources = [
    'src/python/core.c',
    'src/liburash/io.c',
    'src/liburash/internal.c',
    'src/liburash/sha3.c']
if os.name == 'nt':
    sources += [
        'src/liburash/util_win32.c',
        'src/liburash/io_win32.c',
        'src/liburash/mmap_win32.c',
    ]
else:
    sources += [
        'src/liburash/io_posix.c'
    ]
depends = [
    'src/liburash/urash.h',
    'src/liburash/compiler.h',
    'src/liburash/data_sizes.h',
    'src/liburash/endian.h',
    'src/liburash/urash.h',
    'src/liburash/io.h',
    'src/liburash/fnv.h',
    'src/liburash/internal.h',
    'src/liburash/sha3.h',
    'src/liburash/util.h',
]
pyurash = Extension('pyurash',
                     sources=sources,
                     depends=depends,
                     extra_compile_args=["-Isrc/", "-std=gnu99", "-Wall"])

setup(
    name='pyurash',
    author="Matthew Wampler-Doty",
    author_email="matthew.wampler.doty@gmail.com",
    license='GPL',
    version='0.1.23',
    url='https://github.com/ur/urash',
    download_url='https://github.com/ur/urash/tarball/v23',
    description=('Python wrappers for urash, the ur proof of work'
                 'hashing function'),
    ext_modules=[pyurash],
)
