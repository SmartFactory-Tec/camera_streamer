#include "callbacks.h"

void connectSignalHandler(char * signalName, GstElement *element, void *callback, long index){
	g_signal_connect(element, signalName, G_CALLBACK(callback), (gpointer) index);
}
