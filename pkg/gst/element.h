#ifndef ELEMENT_H
#include <gst/gst.h>
#include <stdbool.h>

char *getGstElementName(GstElement *element);

void gst_set_string_property(GstElement *element, char *property_name, char *property_value);

void gst_set_bool_property(GstElement *element, char *property_name, bool *property_value);

void gst_set_int_property(GstElement *element, char *property_name, int property_value);

void gst_set_caps_property(GstElement *element, char *property_name, GstCaps *property_value);

bool setStatePlaying(GstElement* element);
bool setStateNull(GstElement* element);
bool setStatePaused(GstElement* element);
bool setStateReady(GstElement* element);

#endif // PROPERTY_HELPERS_H



