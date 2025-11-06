# Script to generate codebase.md from the codebase using CodeWeaver
# https://github.com/tesserato/codeweaver

codeweaver -clipboard `
    -input "src" `
    -excluded-paths-file "codebase_excluded_paths.txt" `
    -ignore "public\\.+,\.git.*,.+\.exe,.+\.[Pp][Nn][Gg],.+\.[Jj][Pp][Ee]?[Gg],.+\.[Mm][Pp][34],.+\.[Aa][Vv][Ii],.+\.[Mm][Oo][Vv],.+\.[Ww][Mm][Aa],.+\.[Pp][Dd][Ff],.+\.[Dd][Oo][Cc][Xx]?,.+\.[Xx][Ll][Ss][Xx]?,.+\.[Pp][Pp][Tt][Xx]?,.+\.[Zz][Ii][Pp],.+\.[Rr][Aa][Rr],.+\.[7][Zz],.+\.[Ii][Ss][Oo],.+\.[Bb][Ii][Nn],.+\.[Dd][Aa][Tt],.+\.[Dd][Mm][Gg],.+\.[Gg][Ii][Ff],.+\.[Tt][Ii][Ff][Ff]?,.+\.[Pp][Ss][Dd],.+\.[Mm][Kk][Vv],.+\.[Ff][Ll][Aa][Cc],.+\.[Oo][Gg][Gg],.+\.[Ww][Ee][Bb][Pp],.+\.[Aa][Vv][Ii][Ff],.+\.[Hh][Ee][Ii][Cc],.+\.[Jj][Xx][Ll],codebase.md,DRAFT.md,.+\.[Ii][Cc][Oo],.+\.svg"