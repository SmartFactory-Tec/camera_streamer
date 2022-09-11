#include "callbacks.h"

void connectSignalHandler(char * signalName, GstElement *element, void *callback, long index){
	g_signal_connect(element, signalName, G_CALLBACK(callback), (gpointer) index);
}


void callSignalByName(GstElement *element, const char *signalName, void *returnLocation) {
    g_signal_emit_by_name((GObject*)element, signalName, returnLocation);
}
