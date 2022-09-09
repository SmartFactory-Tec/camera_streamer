#ifndef ELEMENT_H
#include <gst/gst.h>
#include <stdbool.h>

extern void padAddedCallback(GstElement *src, GstPad *pad, void *data);

char *getGstElementName(GstElement *element);

void gst_set_string_property(GstElement *element, char *property_name, char *property_value);

void connectPadAdded(GstElement *element);

bool setStatePlaying(GstElement* element);

#endif // PROPERTY_HELPERS_H



