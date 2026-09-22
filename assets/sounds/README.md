# River Raid Sound Placeholders

You can replace any of the built-in procedural synthesizer sounds with your own custom `.wav` audio files!

Just place your audio files in this directory with the following filenames:

- `shoot.wav` - Player blaster / cannon firing sound
- `explosion.wav` - Standard enemy explosion sound
- `big_explosion.wav` - Bridge destruction / massive blast sound
- `fuel.wav` - Fuel depot refuel pickup sound
- `low_fuel.wav` - Low fuel emergency alarm warning
- `extra_life.wav` - Bonus life fanfare
- `engine.wav` - Jet engine jet thruster sound

If any of these files are not present, the game automatically generates dynamic procedural audio in memory using the built-in pure Go wave synthesizer.
