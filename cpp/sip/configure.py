import os
import sipconfig
import subprocess

# The name of the SIP build file generated by SIP and used by the build
# system.

# Get the SIP configuration information.
config = sipconfig.Configuration()

modules = ["node","parser_generator","stringparser"]

sip_dir = os.path.dirname(os.path.abspath(__file__))
src_dir = os.path.join(os.path.dirname(sip_dir),'src')
lib_dir = os.path.join(os.path.dirname(sip_dir),'build')
build_dir = os.path.join(sip_dir,'build')

module_dirs = []
for module in modules:
    # Run SIP to generate the code.
    module_with_slashes = module.replace('.','/')
    module_dir = os.path.join(sip_dir,module_with_slashes)
    module_build_dir = os.path.join(build_dir,module_with_slashes)
    module_dirs.append(os.path.relpath(module_build_dir,sip_dir))
    if not os.path.exists(module_build_dir):
        os.makedirs(module_build_dir)

    module_name = module.split(".")[-1]
    build_file = "build.sbf"
    subprocess.check_output([config.sip_bin,
                             "-c",
                             module_build_dir,
                             "-b",
                             os.path.join(module_build_dir,build_file),
                             "{}.sip".format(module_name)],cwd = module_dir)

    # Create the Makefile.
    makefile = sipconfig.SIPModuleMakefile(config,
                                           build_file,
                                           dir=module_build_dir,
                                           install_dir = '/parsejoy')
    makefile.extra_cxxflags = ['-std=c++14']
    # Add the library we are wrapping.  The name doesn't include any platform
    # specific prefixes or extensions (e.g. the "lib" prefix on UNIX, or the
    # ".dll" extension on Windows).
    makefile.extra_libs = ["parsejoy"]
    makefile.extra_lib_dirs = [os.path.relpath(lib_dir,module_build_dir)]
    makefile.extra_include_dirs = [os.path.relpath(src_dir,module_build_dir)]

    # Generate the Makefile itself.
    makefile.generate()

    sipconfig.ParentMakefile(
        configuration=config,
        dir="",
        subdirs=module_dirs
    ).generate()
