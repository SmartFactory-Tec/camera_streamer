#ifndef ELEMENT_H
#include <gst/gst.h>
#include <stdbool.h>

bool setStatePlaying(GstElement* element);
bool setStateNull(GstElement* element);
bool setStatePaused(GstElement* element);
bool setStateReady(GstElement* element);

#endif // ELEMENT_H



