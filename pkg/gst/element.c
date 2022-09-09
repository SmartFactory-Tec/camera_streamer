#include "element.h"

bool setStatePlaying(GstElement* element) {
	return gst_element_set_state(element, GST_STATE_PLAYING) != GST_STATE_CHANGE_FAILURE;
}

char *getGstElementName(GstElement *element){
    return GST_ELEMENT_NAME(element);
}

void gst_set_string_property(GstElement *element, char *property_name, char *property_value) {
    g_object_set(element, property_name, property_value, NULL);
}

