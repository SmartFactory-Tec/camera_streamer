#ifndef CALLBACK_H
#include <gst/gst.h>


extern void overrunHandler(GstElement* object, long index);

void connectSignalHandler(char *signalName, GstElement *element, void *callback, long index);

void callSignalByName(GstElement *element, const char *signalName, void *returnLocation);

#endif // CALLBACKS_h
