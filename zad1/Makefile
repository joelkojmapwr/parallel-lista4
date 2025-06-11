all: mutex_template

mutex_template: mutex_template.adb
	gnatmake -g mutex_template.adb
	gnatbind mutex_template
	gnatlink mutex_template

run: mutex_template
	./mutex_template > out1

show:
	bash ../PW/bash-tools/display-travel-2.bash out1 

clean:
	rm -f *.o *.ali mutex_template
