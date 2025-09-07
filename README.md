# rename-xattr

Command utility to rename extended file attributes (xattr) maintaining the contents.

This utility was born because I have an old Python utility that creates extended attributes with the hashes of the contents of files using names as `same-hash.md5` and `same-hash.sha512` and now is recommended to use `user.same-hash.md5` and `user.same-hash.sha512` and being normally big files (and many of them) I don't want to recalculate all these hashes/attributes.

## Usage:

`rename-xattr --options files_or_directories ...`

### Options:

+ `--xattr XATTR_NAME`
+ `-X XATTR_NAME`
   Name to rename to the attributes
+ `--source-xattr XATTR_NAME`
+ `-S XATTR_NAME`
   Name to rename from the attributes
+ `--recursive`
+ `-r`
   If to recurse into directories
+ `--verbose`
   Verbose mode
+ `--help`
   Help

It works in MacOS with HFS+, APFS, and (Tuxera) NTFS.
It works on Linux with EXT4 and XFS

R.
