# Script to generate codebase.md from the codebase using CodeWeaver
# https://github.com/tesserato/codeweaver

$instruction = @"
**Before generating any code, state the features added and the features removed (hopefully none) by the code you generated - when in doubt, output this part and ask for confirmation before generating code!**

1. When implementing the requested changes, generate the complete modified files for easy copy and paste;
2. Change the code as little as possible;
3. Do not Introduce regressions or arbitrary simplifications: keep comments, checks, asserts, etc;
4. Generate professional and standard code
5. Do not add ephemerous comments, like `Changed`, `Fix Start`, `Removed`, etc. Always generate a final, professional version of the codebase;
6. Do not add the path at the top of the file.
"@

codeweaver -clipboard `
    -instruction $instruction `
    -include "^src,main.go,go.mod,README.md" `
    -ignore "public\\.+,\.git.*,.+\.exe,.+\.[Pp][Nn][Gg],.+\.[Jj][Pp][Ee]?[Gg],.+\.[Mm][Pp][34],.+\.[Aa][Vv][Ii],.+\.[Mm][Oo][Vv],.+\.[Ww][Mm][Aa],.+\.[Pp][Dd][Ff],.+\.[Dd][Oo][Cc][Xx]?,.+\.[Xx][Ll][Ss][Xx]?,.+\.[Pp][Pp][Tt][Xx]?,.+\.[Zz][Ii][Pp],.+\.[Rr][Aa][Rr],.+\.[7][Zz],.+\.[Ii][Ss][Oo],.+\.[Bb][Ii][Nn],.+\.[Dd][Aa][Tt],.+\.[Dd][Mm][Gg],.+\.[Gg][Ii][Ff],.+\.[Tt][Ii][Ff][Ff]?,.+\.[Pp][Ss][Dd],.+\.[Mm][Kk][Vv],.+\.[Ff][Ll][Aa][Cc],.+\.[Oo][Gg][Gg],.+\.[Ww][Ee][Bb][Pp],.+\.[Aa][Vv][Ii][Ff],.+\.[Hh][Ee][Ii][Cc],.+\.[Jj][Xx][Ll],codebase.md,DRAFT.md,.+\.[Ii][Cc][Oo],.+\.svg" `
    -excluded-paths-file "codebase_excluded_paths.txt"
