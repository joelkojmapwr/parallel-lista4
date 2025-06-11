all: mutex_template

mutex_template: mutex_template.adb
	gnatmake -g  -gnato -gnatE -gnatf -gnatVa mutex_template.adb


run: mutex_template
	./mutex_template > out1

show:
	bash ../PW/bash-tools/display-travel-2.bash out1 

clean:
	rm -f *.o *.ali mutex_template b~*
