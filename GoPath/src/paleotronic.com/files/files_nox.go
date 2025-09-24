// +build nox

package files

var (
	dskp0       *DSKFileProvider = NewDSKFileProvider("", 0)
	dskp1       *DSKFileProvider = NewDSKFileProvider("", 0)
	r_providers []FileProvider   = []FileProvider{
		NewMappedFileProvider(
			map[string]ProviderHolder{
				"media":     ProviderHolder{BasePath: "", Provider: NewNetworkUserFileProvider("media", false, true, 0)},
				"micropaks": ProviderHolder{BasePath: "", Provider: NewNetworkUserFileProvider("micropaks", false, true, 0)},
				"appleii":   ProviderHolder{BasePath: "", Provider: NewNetworkUserFileProvider("", false, true, 0)},
				"spectrum":  ProviderHolder{BasePath: "", Provider: NewNetworkUserFileProvider("spectrum", false, true, 0)},
				"system":    ProviderHolder{BasePath: "", Provider: NewNetworkSystemFileProvider("", true, false, 0)},
				"disk0":     ProviderHolder{BasePath: "", Provider: dskp0},
				"disk1":     ProviderHolder{BasePath: "", Provider: dskp1},
				"":          ProviderHolder{BasePath: "", Provider: NewNetworkRemIntFileProvider("", "uuid", true, false, 0)},
			}),
	}
	s_providers []FileProvider = []FileProvider{
		NewMappedFileProvider(
			map[string]ProviderHolder{
				"":     ProviderHolder{BasePath: "", Provider: NewLocalFileProvider(GetBinPath(), true, 0)},
				"fs":   ProviderHolder{BasePath: "", Provider: NewLocalFileProvider(GetSysRoot(), false, 0)},
				"boot": ProviderHolder{BasePath: "", Provider: NewInternalFileProvider("bootsystem/boot", 0)},
			}),
	}
	e_providers []FileProvider = []FileProvider{
		NewMappedFileProvider(
			map[string]ProviderHolder{
				"boot":   ProviderHolder{BasePath: "", Provider: NewInternalFileProvider("bootsystem/boot", 0)},
				"system": ProviderHolder{BasePath: "", Provider: NewInternalFileProvider("eboot/system", 0)},
				"local":  ProviderHolder{BasePath: "", Provider: NewLocalFileProvider(GetUserDirectory(BASEDIR), true, 0)},
				"disk0":  ProviderHolder{BasePath: "", Provider: dskp0},
				"disk1":  ProviderHolder{BasePath: "", Provider: dskp1},
				"fs":     ProviderHolder{BasePath: "", Provider: NewLocalFileProvider(GetSysRoot(), false, 0)},
			}),
	}
	u_providers []FileProvider = []FileProvider{
		NewMappedFileProvider(
			map[string]ProviderHolder{
				"":     ProviderHolder{BasePath: "", Provider: NewLocalFileProvider(GetBinPath(), true, 0)},
				"fs":   ProviderHolder{BasePath: "", Provider: NewLocalFileProvider(GetSysRoot(), false, 0)},
				"boot": ProviderHolder{BasePath: "", Provider: NewInternalFileProvider("bootsystem/boot", 0)},
			}),
	}
	p_providers []FileProvider = []FileProvider{
		NewMappedFileProvider(
			map[string]ProviderHolder{"userdata": ProviderHolder{BasePath: "pigfrog", Provider: NewNetworkUserFileProvider("", true, false, 0)},
				"": ProviderHolder{BasePath: "", Provider: NewProjectProvider("", true, true, true, "pigfrog", 0)},
			}),
	}
)
