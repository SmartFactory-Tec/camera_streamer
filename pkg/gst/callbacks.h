#ifndef CALLBACK_H
#include <gst/gst.h>

extern void padAddedHandler(GstElement*, GstPad* newPad, long index);

void connectSignalHandler(char *signalName, GstElement *element, void *callback, long index);

#endif // CALLBACKS_h
