Output demo.gif

Set FontSize 20
Set Width 1280
Set Height 720
Set WindowBar Colorful
Set Padding 10

Hide
Type "./logviewer test --config config.template.toml"
Enter
Sleep 200ms
Show

# home screen

Sleep 1500ms

# select a pod

Right@500ms 2

Down@500ms 1

# slect a container and show the logs

Right@500ms 2

Sleep 300ms

Enter

Sleep 1500ms

# enter in the return fields view

Type r

Sleep 500ms

Type "level, msg, bool_field, int_field"

Ctrl+R

Sleep 1s

Escape

Sleep 500ms

# enter in the filters view

Type f

Sleep 500ms

Type 'json.bool_field and json.msg contains "test"'

Ctrl+R

Sleep 500ms

Escape

Sleep 2s

Ctrl+C